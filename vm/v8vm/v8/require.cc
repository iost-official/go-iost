#include "require.h"

//#include "esprima.js.h"
//#include "inject_gas.js.h"
#include "storage.js.h"
#include "blockchain.js.h"

#include <stdlib.h>
#include <fstream>
#include <sstream>
#include <iostream>
#include <unordered_map>

std::unordered_map<std::string, const char *> libJS = {
//    {"esprima", reinterpret_cast<char *>(__libjs_esprima_js)},
//    {"inject_gas", reinterpret_cast<char *>(__libjs_inject_gas_js)},
    {"storage", reinterpret_cast<char *>(__libjs_storage_js)},
    {"blockchain", reinterpret_cast<char *>(__libjs_blockchain_js)}
};

static requireFunc CRequire = nullptr;

void InitGoRequire(requireFunc require) {
    CRequire = require;
}

void NewNativeRequire(const FunctionCallbackInfo<Value> &info) {
    Isolate *isolate = info.GetIsolate();
    Local<Context> context = isolate->GetCurrentContext();
    Local<Object> global = context->Global();
    Local<Value> val = global->GetInternalField(0);
    if (!val->IsExternal()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "nativeRequire val error")
        );
        isolate->ThrowException(err);
        return;
    }
    SandboxPtr sbxPtr = static_cast<SandboxPtr>(Local<External>::Cast(val)->Value());
    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);

    Local<Value> path = info[0];
    if (!path->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "require empty module")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value pathStr(path);

    if (libJS.find(*pathStr) == libJS.end()) {
        return;
    }

    // if it's jsFile under jsPath
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, libJS[*pathStr]));
    return;
}

void InitRequire(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    globalTpl->Set(
        String::NewFromUtf8(isolate, "_native_require", NewStringType::kNormal)
                      .ToLocalChecked(),
        FunctionTemplate::New(isolate, NewNativeRequire));
}
