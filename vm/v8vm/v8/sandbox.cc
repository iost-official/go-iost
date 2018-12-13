#include "sandbox.h"
#include "console.h"
#include "require.h"
#include "storage.h"
#include "blockchain.h"
#include "instruction.h"
#include "crypto.h"

#include "vm.js.h"
#include "compile_vm.js.h"
#include "environment.js.h"
#include <assert.h>
#include <cstring>
#include <fstream>
#include <sstream>
#include <thread>
#include <stdlib.h>
#include <stdio.h>
#include <thread>
#include <iostream>
#include <unistd.h>
#include <chrono>

char *vmJsLib = reinterpret_cast<char *>(__libjs_vm_js);
char *envJsLib = reinterpret_cast<char *>(__libjs_environment_js);
char *compileVmJsLib = reinterpret_cast<char *>(__libjs_compile_vm_js);

const char *preloadBlockCode = R"(
// load Block
const blockInfo = JSON.parse(blockchain.blockInfo());
const block = {
   number: blockInfo.number,
   parentHash: blockInfo.parent_hash,
   witness: blockInfo.witness,
   time: blockInfo.time
};


// load tx
const txInfo = JSON.parse(blockchain.txInfo());
const tx = {
   time: txInfo.time,
   hash: txInfo.hash,
   expiration: txInfo.expiration,
   gasLimit: txInfo.gas_limit,
   gasRatio: txInfo.gas_ratio,
   authList: txInfo.auth_list,
   publisher: txInfo.publisher
};)";

const int sandboxMemLimit = 100000000; // 100mb
void copyString(CStr &cstr, const std::string &str) {
    cstr.size = str.length();
    cstr.data = new char[cstr.size + 1];
    std::memcpy(cstr.data, str.c_str(), cstr.size + 1);
    return;
}

std::string v8ValueToStdString(Local<Value> val) {
    String::Utf8Value str(val);
    if (str.length() == 0) {
        return "";
    }
    return *str;
}

void nativeLog(const FunctionCallbackInfo<Value> &info) {
    Isolate *isolate = info.GetIsolate();

    Local<Value> msg = info[0];
    if (!msg->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "_native_log empty log")
        );
        isolate->ThrowException(err);
    }

    String::Utf8Value msgStr(msg);
    std::cout << "native_log: " << *msgStr << std::endl;
    return;
}

void nativeRun(const FunctionCallbackInfo<Value> &info) {
    Isolate *isolate = info.GetIsolate();

    Local<Value> source = info[0];
    Local<Value> fileName = info[1];
    if (!fileName->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "_native_run empty script.")
        );
        isolate->ThrowException(err);
    }

    Local<String> source2 = String::NewFromUtf8(isolate, v8ValueToStdString(source).c_str(), NewStringType::kNormal).ToLocalChecked();
    Local<String> fileName2 = String::NewFromUtf8(isolate, v8ValueToStdString(fileName).c_str(), NewStringType::kNormal).ToLocalChecked();
    Local<Script> script = Script::Compile(source2, fileName2);

    if (!script.IsEmpty()) {
        Local<Value> result = script->Run();
        if (!result.IsEmpty()) {
            info.GetReturnValue().Set(result);
        }
    }

    return;
}

Local<ObjectTemplate> createGlobalTpl(Isolate *isolate) {
    Local<ObjectTemplate> global = ObjectTemplate::New(isolate);
    global->SetInternalFieldCount(1);

//    InitConsole(isolate, global);
//    InitRequire(isolate, global);
    InitStorage(isolate, global);
    InitBlockchain(isolate, global);
    InitInstruction(isolate, global);
    InitCrypto(isolate, global);

    global->Set(
              String::NewFromUtf8(isolate, "_native_log", NewStringType::kNormal)
                  .ToLocalChecked(),
              v8::FunctionTemplate::New(isolate, nativeLog));

    global->Set(
                      String::NewFromUtf8(isolate, "_native_run", NewStringType::kNormal)
                          .ToLocalChecked(),
                      v8::FunctionTemplate::New(isolate, nativeRun));

    return global;
}

const char* ToCString(const v8::String::Utf8Value& value) {
  return *value ? *value : "<string conversion failed>";
}

SandboxPtr newSandbox(IsolateWrapperPtr ptr) {
    IsolateWrapper *isolateWrapper = static_cast<IsolateWrapper*>(ptr);
    Isolate *isolate = static_cast<Isolate*>(isolateWrapper->isolate);
    ArrayBufferAllocator* allocator= static_cast<ArrayBufferAllocator*>(isolateWrapper->allocator);

    Locker locker(isolate);
    Isolate::Scope isolate_scope(isolate);
    HandleScope handle_scope(isolate);

    Local<ObjectTemplate> globalTpl = createGlobalTpl(isolate);
    Local<Context> context = Context::New(isolate, NULL, globalTpl);
    context->AllowCodeGenerationFromStrings(false);

    Sandbox *sbx = new Sandbox();
    Local<Object> global = context->Global();
    global->SetInternalField(0, External::New(isolate, sbx));

    sbx->context.Reset(isolate, context);
    sbx->isolate = isolate;
    sbx->allocator = allocator;
    sbx->jsPath = strdup("v8/libjs");
    sbx->gasUsed = 0;
    sbx->gasLimit = 0;
    sbx->memLimit = sandboxMemLimit;

    return static_cast<SandboxPtr>(sbx);
}

void releaseSandbox(SandboxPtr ptr) {
    if (ptr == nullptr) {
        return;
    }

    Sandbox *sbx = static_cast<Sandbox*>(ptr);

    sbx->context.Reset();

    free((char *)sbx->jsPath);
    delete sbx;
    return;
}

void setJSPath(SandboxPtr ptr, const char *jsPath) {
    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    if (sbx->jsPath != nullptr)
        free((char *)sbx->jsPath);
    sbx->jsPath = strdup(jsPath);
}

void setSandboxGasLimit(SandboxPtr ptr, size_t gasLimit) {
    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    sbx->gasLimit = gasLimit;
}

void setSandboxMemLimit(SandboxPtr ptr, size_t memLimit) {
    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    sbx->memLimit = memLimit;
}

std::string reportException(Isolate *isolate, Local<Context> ctx, TryCatch& tryCatch) {
    std::stringstream ss;
    ss << "Uncaught exception: ";

    if (tryCatch.Message().IsEmpty()) {
        return ss.str();
    }

    ss << v8ValueToStdString(tryCatch.Exception());

    if (!tryCatch.Message().IsEmpty()) {
        if (!tryCatch.Message()->GetScriptResourceName()->IsUndefined()) {
            ss << std::endl;
            ss << "at " << v8ValueToStdString(tryCatch.Message()->GetScriptResourceName());

            Maybe<int> lineNo = tryCatch.Message()->GetLineNumber(ctx);
            Maybe<int> start = tryCatch.Message()->GetStartColumn(ctx);
            Maybe<int> end = tryCatch.Message()->GetEndColumn(ctx);
            MaybeLocal<String> sourceLine = tryCatch.Message()->GetSourceLine(ctx);

            if (lineNo.IsJust()) {
                ss << ":" << lineNo.ToChecked();
            }
            if (start.IsJust()) {
                ss << ":" << start.ToChecked();
            }
            if (!sourceLine.IsEmpty()) {
                ss << std::endl;
                ss << "  " << v8ValueToStdString(sourceLine.ToLocalChecked());
            }
            if (start.IsJust() && end.IsJust()) {
                ss << std::endl;
                ss << "  ";
                for (int i = 0; i < start.ToChecked(); i++) {
                    ss << " ";
                }
                for (int i = start.ToChecked(); i < end.ToChecked(); i++) {
                    ss << "^";
                }
            }
        }
    }

    if (!tryCatch.StackTrace().IsEmpty()) {
        ss << std::endl;
        ss << "Stack tree: " << std::endl;
        ss << v8ValueToStdString(tryCatch.StackTrace());
    }

    return ss.str();
}

void loadVM(SandboxPtr ptr, int vmType) {
    if (ptr == nullptr) {
        return;
    }

    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    Isolate *isolate = sbx->isolate;
    Locker locker(isolate);
    Isolate::Scope isolate_scope(isolate);
    HandleScope handle_scope(isolate);

    Local<Context> context = sbx->context.Get(isolate);
    Context::Scope context_scope(context);

    if (vmType == 0) {
        return;
    }

    // load vm
    const char *vmPath = "vm.js";
    Local<String> source = String::NewFromUtf8(isolate, vmJsLib, NewStringType::kNormal).ToLocalChecked();
    Local<String> fileName = String::NewFromUtf8(isolate, vmPath, NewStringType::kNormal).ToLocalChecked();
    Local<Script> script = Script::Compile(source, fileName);
    if (!script.IsEmpty()) {
        Local<Value> result = script->Run();
    }

    // load environment
    const char *envPath = "env.js";
    source = String::NewFromUtf8(isolate, envJsLib, NewStringType::kNormal).ToLocalChecked();

    fileName = String::NewFromUtf8(isolate, envPath, NewStringType::kNormal).ToLocalChecked();
    script = Script::Compile(source, fileName);
    if (!script.IsEmpty()) {
        Local<Value> result = script->Run();
    }
}

size_t MemoryUsage(Isolate* isolate, ArrayBufferAllocator* allocator) {
    // V8 memory usage
    HeapStatistics v8_heap_stats;
    isolate->GetHeapStatistics(&v8_heap_stats);

    /*fields[1] = v8_heap_stats.total_heap_size();
    fields[2] = v8_heap_stats.used_heap_size();
    fields[3] = v8_heap_stats.external_memory();*/
    //int a = v8_heap_stats.total_heap_size() + allocator->GetMaxAllocatedMemSize();
    //std::cout << "MemoryUsed: " << a - startMemHHH  << std::endl;
    return v8_heap_stats.total_heap_size() + allocator->GetMaxAllocatedMemSize();
}

void RealExecute(SandboxPtr ptr, const CStr code, std::string &result, std::string &error, bool &isJson, bool &isDone) {
    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    Isolate *isolate = sbx->isolate;

    Locker locker(isolate);
    Isolate::Scope isolate_scope(isolate);
    HandleScope handle_scope(isolate);
    Local<Context> context = sbx->context.Get(isolate);
    Context::Scope context_scope(context);

    TryCatch tryCatch(isolate);
    tryCatch.SetVerbose(true);

    // preload block info.
    Local<String> source = String::NewFromUtf8(isolate, preloadBlockCode, NewStringType::kNormal).ToLocalChecked();
    Local<String> fileName = String::NewFromUtf8(isolate, "_preload_block.js", NewStringType::kNormal).ToLocalChecked();
    Local<Script> script = Script::Compile(source, fileName);

    Local<Value> ret = script->Run();
    if (tryCatch.HasCaught() && !tryCatch.Exception()->IsNull()) {
        std::string exception = reportException(isolate, context, tryCatch);
        error = exception;
        return;
    }

    // reset gas count
    sbx->gasUsed = 0;

    source = String::NewFromUtf8(isolate, code.data, NewStringType::kNormal, code.size).ToLocalChecked();
    fileName = String::NewFromUtf8(isolate, "_default_name.js", NewStringType::kNormal).ToLocalChecked();
    script = Script::Compile(source, fileName);

    if (script.IsEmpty()) {
        std::string exception = reportException(isolate, context, tryCatch);
        error = exception;
        return;
    }

    ret = script->Run();

    if (tryCatch.HasCaught() && tryCatch.Exception()->IsNull()) {
        isDone = true;
        return;
    }

    if (tryCatch.HasCaught() && !tryCatch.Exception()->IsNull()) {
        std::string exception = reportException(isolate, context, tryCatch);
        error = exception;
        return;
    }

    if (ret->IsString() || ret->IsNumber() || ret->IsBoolean()) {
        String::Utf8Value retV8Str(isolate, ret);
        result.assign(*retV8Str, retV8Str.length());
        isDone = true;
        return;
    }

    Local<Object> obj = ret.As<Object>();
    if (!obj->IsUndefined()) {
        MaybeLocal<String> jsonRet = JSON::Stringify(context, obj);
        if (!jsonRet.IsEmpty()) {
            isJson = true;
            String::Utf8Value jsonRetStr(jsonRet.ToLocalChecked());
            result.assign(*jsonRetStr, jsonRetStr.length());
        }

        if (tryCatch.HasCaught() && !tryCatch.Exception()->IsNull()) {
            std::string exception = reportException(isolate, context, tryCatch);
            error = exception;
            return;
        }
    }
    isDone = true;
    return;
}

ValueTuple Execution(SandboxPtr ptr, const CStr code, long long int expireTime) {
    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    Isolate *isolate = sbx->isolate;

    std::string result;
    std::string error;
    bool isJson = false;
    bool isDone = false;
    //std::cout << "StartMemBeforeChange: " << startMemHHH << std::endl;
    //startMemHHH = MemoryUsage(isolate, sbx->allocator);
    std::thread exec(RealExecute, ptr, code, std::ref(result), std::ref(error), std::ref(isJson), std::ref(isDone));

    ValueTuple res = { {nullptr, 0}, {nullptr, 0}, isJson, 0 };
//    auto startTime = std::chrono::steady_clock::now();
    while(true) {
        if (error.length() > 0) {
            copyString(res.Err, error);
            res.gasUsed = sbx->gasUsed;
            break;
        }
        if (result.length() > 0) {
            copyString(res.Value, result);
            res.isJson = isJson;
            res.gasUsed = sbx->gasUsed;
            break;
        }
        if (isDone) {
            copyString(res.Value, result);
            res.isJson = isJson;
            res.gasUsed = sbx->gasUsed;
            break;
        }
  /*      if (MemoryUsage(isolate, sbx->allocator) > sbx->memLimit) {
            isolate->TerminateExecution();
            copyString(res.Err, "out of memory");
            res.gasUsed = sbx->gasLimit;
            break;
        } */
        if (sbx->gasUsed > sbx->gasLimit) {
            isolate->TerminateExecution();
            copyString(res.Err, "out of gas");
            res.gasUsed = sbx->gasUsed;
            break;
        }
        auto now = std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::system_clock::now().time_since_epoch()).count();
        //auto execTime = std::chrono::duration_cast<std::chrono::milliseconds>(now - startTime).count();
        if (now > expireTime) {
            isolate->TerminateExecution();
            copyString(res.Err, ("execution killed, current time : " + std::to_string(now) + " , expireTime: " + std::to_string(expireTime)).c_str());
            res.gasUsed = sbx->gasUsed;
            break;
        }
        //usleep(10);
        std::this_thread::sleep_for(std::chrono::microseconds(10));
    }
    if (exec.joinable())
        exec.join();
    //std::cout << " MemoryUsed: " << MemoryUsage(isolate, sbx->allocator) - startMemHHH << std::endl;
    return res;
}
