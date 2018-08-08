#include "require.h"

#include <stdlib.h>
#include <fstream>
#include <sstream>

using namespace v8;

#define NATIVE_LIB_PATH "v8/libjs/"

//static int readFile(const char *filename, char **code) {
//    Isolate *isolate = info.GetIsolate();
//
//    Local<Value> fileName = info[0];
//    if (!fileName->IsString()) {
//        Local<Value> err = Exception::Error(
//            String::NewFromUtf8(isolate, "readFile empty file name.")
//        );
//        isolate->ThrowException(err);
//    }
//
//    String::Utf8Value fileNameStr(fileName);
//    std::string fullRelPath = std::string(NATIVE_LIB_PATH) + *fileNameStr + std::string(".js");
//
//    std::ifstream f(fullRelPath);
//
//    if (f.good()) {
        // file not found
//        return 1;
//    }
//
//    std::stringstream buffer;
//    buffer << f.rdbuf();
//
//    info.GetReturnValue().Set(String::NewFromUtf8(isolate, buffer.str().c_str()));
//
//    return 0;
//}

void nativeRequire(const FunctionCallbackInfo<Value> &info) {
    Isolate *isolate = info.GetIsolate();

    Local<Value> path = info[0];
    if (!path->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "require empty module")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value pathStr(path);
    std::string fullRelPath = std::string(NATIVE_LIB_PATH) + *pathStr + std::string(".js");

    std::ifstream f(fullRelPath);
    std::stringstream buffer;
    buffer << f.rdbuf();

    if (buffer.str().length() > 0) {
        info.GetReturnValue().Set(String::NewFromUtf8(isolate, buffer.str().c_str()));
        return;
    }

    // read go module again
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
    SandboxPtr sbx = static_cast<SandboxPtr>(Local<External>::Cast(val)->Value());

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
