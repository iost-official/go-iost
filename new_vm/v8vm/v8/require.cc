#include "require.h"

#include <iostream>
#include <stdlib.h>
#include <iostream>

using namespace v8;

void nativeRequire(const FunctionCallbackInfo<Value> &info) {
    Isolate *isolate = info.GetIsolate();
    Local<Context> context = isolate->GetCurrentContext();
    Local<Object> global = context->Global();

    Local<Value> val = global->GetInternalField(0);
    if (!val->IsExternal()) {
           std::cout << "nativeRequire val error" << std::endl;
        return;
    }

    SandboxPtr sbx = static_cast<SandboxPtr>(Local<External>::Cast(val)->Value());

    Local<Value> path = info[0];
    if (!path->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "_native_require empty path")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value pathStr(path);
    char *code = requireModule(sbx, *pathStr);

    info.GetReturnValue().Set(String::NewFromUtf8(isolate, code));
    free(code);
}

void InitRequire(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    globalTpl->Set(
        String::NewFromUtf8(isolate, "_native_require", NewStringType::kNormal)
                      .ToLocalChecked(),
        FunctionTemplate::New(isolate, nativeRequire));
}
