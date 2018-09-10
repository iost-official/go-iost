#include "require.h"
#include "bignumber.js.h"
#include "blockchain.js.h"
#include "console.js.h"
#include "esprima.js.h"
#include "inject_gas.js.h"
#include "int64.js.h"
#include "storage.js.h"
#include "utils.js.h"
#include "vm.js.h"

#include <stdlib.h>
#include <fstream>
#include <sstream>
#include <iostream>
#include <unordered_map>

std::unordered_map<std::string, const char *> jsLib = {
    {"bignumber", reinterpret_cast<char *>(__libjs_bignumber_js)},
    {"blockchain", reinterpret_cast<char *>(__libjs_blockchain_js)},
    {"console", reinterpret_cast<char *>(__libjs_console_js)},
    {"esprima", reinterpret_cast<char *>(__libjs_esprima_js)},
    {"inject_gas", reinterpret_cast<char *>(__libjs_inject_gas_js)},
    {"int64", reinterpret_cast<char *>(__libjs_int64_js)},
    {"storage", reinterpret_cast<char *>(__libjs_storage_js)},
    {"utils", reinterpret_cast<char *>(__libjs_utils_js)},
    {"vm", reinterpret_cast<char *>(__libjs_vm_js)},
};

static char injectGasFormat[] =
    "(function(){\n"
    "const source = \"%s\";\n"
    "return injectGas(source);\n"
    "})();";
static requireFunc CRequire = nullptr;

void InitGoRequire(requireFunc require) {
    CRequire = require;
}

void nativeRequire(const FunctionCallbackInfo<Value> &info) {
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
    std::string fullRelPath = std::string("__libjs_") + *pathStr + "_js";

    if (jsLib.find(fullRelPath) == jsLib.end()) {
        return;
    }

    std::cout << jsLib[fullRelPath] << std::endl;
    Local<String> jsLibStr = String::NewFromUtf8(isolate, jsLib[fullRelPath], NewStringType::kNormal).ToLocalChecked();
    info.GetReturnValue().Set(jsLibStr);
    return;
}

void InitRequire(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    globalTpl->Set(
        String::NewFromUtf8(isolate, "_native_require", NewStringType::kNormal)
                      .ToLocalChecked(),
        FunctionTemplate::New(isolate, nativeRequire));
}
