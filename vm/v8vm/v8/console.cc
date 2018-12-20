#include "console.h"
#include <iostream>

static consoleFunc CConsole = nullptr;

void InitGoConsole(consoleFunc console) {
    CConsole = console;
}

void NewConsoleLog(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Context> context = isolate->GetCurrentContext();
    Local<Object> global = context->Global();
    Local<Value> val = global->GetInternalField(0);
    if (!val->IsExternal()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "consoleLog val error")
        );
        isolate->ThrowException(err);
        return;
    }
    SandboxPtr sbxPtr = static_cast<SandboxPtr>(Local<External>::Cast(val)->Value());

    if (args.Length() != 2) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "consoleLog invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> levelVal = args[0];
    if (!levelVal->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "consoleLog log level must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> logVal = args[1];
    if (!logVal->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "consoleLog log message must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    NewCStrChecked(levelStr, levelVal, isolate);
    NewCStrChecked(logStr, logVal, isolate);

    CConsole(sbxPtr, levelStr, logStr);
}

void InitConsole(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    globalTpl->Set(
        String::NewFromUtf8(isolate, "_cLog", NewStringType::kNormal)
                    .ToLocalChecked(),
        FunctionTemplate::New(isolate, NewConsoleLog));
}
