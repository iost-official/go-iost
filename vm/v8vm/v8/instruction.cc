#include "instruction.h"
#include <iostream>

void NewIOSTContractInstruction(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Context> context = isolate->GetCurrentContext();
    Local<Object> global = context->Global();

    Local<Value> val = global->GetInternalField(0);
    if (!val->IsExternal()) {
           std::cout << "NewIOSTContractInstruction val error" << std::endl;
        return;
    }
    SandboxPtr sbx = static_cast<SandboxPtr>(Local<External>::Cast(val)->Value());

    IOSTContractInstruction *ici = new IOSTContractInstruction(sbx);

    Local<Object> self = args.Holder();
    self->SetInternalField(0, External::New(isolate, ici));

    args.GetReturnValue().Set(self);
}

void IOSTContractInstruction_Incr(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 1) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractInstruction_Incr invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> val = args[0];
    if (!val->IsNumber()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractInstruction_Incr value must be number")
        );
        isolate->ThrowException(err);
        return;
    }

    int32_t valInt = val->Int32Value();
    if (valInt < 0) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractInstruction_Incr invalid gas")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTContractInstruction_Incr val error" << std::endl;
        return;
    }

    IOSTContractInstruction *ici = static_cast<IOSTContractInstruction *>(extVal->Value());
    size_t ret = ici->Incr(valInt);

    args.GetReturnValue().Set(Number::New(isolate, (double)ret));

    ici->MemUsageCheck();
    return;
}

void IOSTContractInstruction_Count(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 0) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractInstruction_Count invalid argument length.")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTContractInstruction_Count val error" << std::endl;
        return;
    }

    IOSTContractInstruction *ici = static_cast<IOSTContractInstruction *>(extVal->Value());
    size_t ret = ici->Incr(0);

    args.GetReturnValue().Set(Number::New(isolate, (double)ret));
}

void InitInstruction(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    Local<FunctionTemplate> instructionClass =
        FunctionTemplate::New(isolate, NewIOSTContractInstruction);
    Local<String> instructionClassName = String::NewFromUtf8(isolate, "IOSTInstruction");
    instructionClass->SetClassName(instructionClassName);

    Local<ObjectTemplate> instructionTpl = instructionClass->InstanceTemplate();
    instructionTpl->SetInternalFieldCount(1);
    instructionTpl->Set(
        String::NewFromUtf8(isolate, "incr"),
        FunctionTemplate::New(isolate, IOSTContractInstruction_Incr)
    );
    instructionTpl->Set(
        String::NewFromUtf8(isolate, "count"),
        FunctionTemplate::New(isolate, IOSTContractInstruction_Count)
    );

    globalTpl->Set(instructionClassName, instructionClass);
}
