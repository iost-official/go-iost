#include "compile.h"
#include <cstring>

#include "console.h"
#include "require.h"

#include "bignumber.js.h"
#include "int64.js.h"
#include "float64.js.h"
#include "utils.js.h"
#include "console.js.h"
#include "esprima.js.h"
#include "escodegen.js.h"
#include "inject_gas.js.h"
#include "validate.js.h"

intptr_t externalRef[] = {
        reinterpret_cast<intptr_t>(NewConsoleLog),
        reinterpret_cast<intptr_t>(NewNativeRequire),
        0};

static char injectGasFormat[] =
    "(function(){\n"
    "const source = \"%s\";\n"
    "return injectGas(source);\n"
    "})();";

static char validateFormat[] =
    "(function(){\n"
    "const source = \"%s\";\n"
    "return validate(source, %s);\n"
    "})();";

static char codeFormat[] =
        "let module = {};\n"
        "module.exports = {};\n"
        "%s\n"  // load BigNumber
        "let BigNumber = module.exports;\n"
        "%s\n"  // load Int64
        "%s\n"  // load Float64
        "%s\n"  // load util
        "%s\n"; // load console

static char compileCodeFormat[] =
    "let exports = {};\n"
    "let module = {};\n"
    "module.exports = {};\n"
    "%s\n"  // load esprima
    "const esprima = module.exports;\n"
    "%s\n"  // load escodegen
    "const escodegen = module.exports;\n"
    "%s\n"  // load validate
    "%s\n"; // load inject_gas

int compileInternal(SandboxPtr ptr, const CStr code, const CStr extra, char *format, const char *file, CStr *ret, CStr *errMsg) {
    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    Isolate *isolate = sbx->isolate;

    Locker locker(isolate);
    Isolate::Scope isolate_scope(isolate);
    HandleScope handle_scope(isolate);

    Local<Context> context = sbx->context.Get(isolate);
    Context::Scope context_scope(context);

    TryCatch tryCatch(isolate);
    tryCatch.SetVerbose(false);

    char *formatedCode = nullptr;
    // TODO: print with length
    if (extra.data == nullptr) {
        asprintf(&formatedCode, format, code.data);
    } else {
        asprintf(&formatedCode, format, code.data, extra.data);
    }

    Local<String> source = String::NewFromUtf8(isolate, formatedCode, NewStringType::kNormal).ToLocalChecked();
    free(formatedCode);
    Local<String> fileName = String::NewFromUtf8(isolate, file, NewStringType::kNormal).ToLocalChecked();
    Local<Script> script = Script::Compile(source, fileName);

    if (!script.IsEmpty()) {
        Local<Value> result = script->Run();
        if (!result.IsEmpty()) {
            String::Utf8Value retStr(result);
            // inject gas failed will return empty result. todo catch exception
            if (retStr.length() == 0) {
                return 1;
            }
            ret->data = strdup(*retStr);
            ret->size = retStr.length();
            return 0;
        }

        if (tryCatch.HasCaught()) {
            std::string exception = reportException(isolate, context, tryCatch);
            errMsg->data = strdup(exception.c_str());
            errMsg->size = exception.length();
            return 1;
        }
    }
    errMsg->data = strdup("script is empty");
    errMsg->size = std::strlen(errMsg->data);
    return 1;
}

int compile(SandboxPtr ptr, const CStr code, CStr *compiledCode, CStr *errMsg) {
    return compileInternal(ptr, code, {nullptr, 0}, injectGasFormat, "__inject_gas.js", compiledCode, errMsg);
}

int validate(SandboxPtr ptr, const CStr code, const CStr abi, CStr *result, CStr *errMsg) {
    return compileInternal(ptr, code, abi, validateFormat, "__validate.js", result, errMsg);
}

CustomStartupData createStartupData() {
    char *bignumberjs = reinterpret_cast<char *>(__libjs_bignumber_js);
    char *int64js = reinterpret_cast<char *>(__libjs_int64_js);
    char *float64js = reinterpret_cast<char *>(__libjs_float64_js);
    char *utilsjs = reinterpret_cast<char *>(__libjs_utils_js);
    char *consolejs = reinterpret_cast<char *>(__libjs_console_js);

    char *code = nullptr;
    asprintf(&code, codeFormat,
        bignumberjs,
        int64js,
        float64js,
        utilsjs,
        consolejs);

    StartupData blob;
    {
        SnapshotCreator creator(externalRef);
        Isolate* isolate = creator.GetIsolate();
        {
            HandleScope handle_scope(isolate);

            Local<ObjectTemplate> globalTpl = ObjectTemplate::New(isolate);
            globalTpl->SetInternalFieldCount(1);

            // add console log
            Local<FunctionTemplate> callback = FunctionTemplate::New(isolate, NewConsoleLog);
            globalTpl->Set(String::NewFromUtf8(isolate, "_cLog", NewStringType::kNormal).ToLocalChecked(), callback);

            // add require
            callback = FunctionTemplate::New(isolate, NewNativeRequire);
            globalTpl->Set(String::NewFromUtf8(isolate, "_native_require", NewStringType::kNormal).ToLocalChecked(), callback);

            Local<Context> context = Context::New(isolate, nullptr, globalTpl);
            Context::Scope context_scope(context);

            Local<String> source = String::NewFromUtf8(isolate, code, NewStringType::kNormal).ToLocalChecked();
            Local<Script> script = Script::Compile(context, source).ToLocalChecked();
            if (!script.IsEmpty()){
                script->Run();
            }

            creator.SetDefaultContext(context);
        }
        blob = creator.CreateBlob(SnapshotCreator::FunctionCodeHandling::kClear);
    }

    return CustomStartupData{blob.data, blob.raw_size};
}

CustomStartupData createCompileStartupData() {
    char *esprimajs = reinterpret_cast<char *>(__libjs_esprima_js);
    char *escodegenjs = reinterpret_cast<char *>(__libjs_escodegen_js);
    char *validatejs = reinterpret_cast<char *>(__libjs_validate_js);
    char *injectgasjs = reinterpret_cast<char *>(__libjs_inject_gas_js);

    char *code = nullptr;
    asprintf(&code, compileCodeFormat,
        esprimajs,
        escodegenjs,
        validatejs,
        injectgasjs);

    StartupData blob;
    {
        SnapshotCreator creator;
        Isolate* isolate = creator.GetIsolate();
        {
            HandleScope handle_scope(isolate);

            Local<Context> context = Context::New(isolate);
            Context::Scope context_scope(context);

            Local<String> source = String::NewFromUtf8(isolate, code, NewStringType::kNormal).ToLocalChecked();
            Local<Script> script = Script::Compile(context, source).ToLocalChecked();
            if (!script.IsEmpty()){
                script->Run();
            }

            creator.SetDefaultContext(context);
        }
        blob = creator.CreateBlob(SnapshotCreator::FunctionCodeHandling::kClear);
    }

    return CustomStartupData{blob.data, blob.raw_size};
}
