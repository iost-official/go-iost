#include "storage.h"
#include <iostream>

static putFunc CPut = nullptr;
static getFunc CGet = nullptr;
static delFunc CDel = nullptr;
static globalGetFunc CGGet = nullptr;

void InitGoStorage(putFunc put, getFunc get, delFunc del, globalGetFunc gGet) {
    CPut = put;
    CGet = get;
    CDel = del;
    CGGet = gGet;
}

int IOSTContractStorage::Put(const char *key, const char *value) {
    size_t gasUsed = 0;
    int ret = CPut(sbxPtr, key, value, &gasUsed);
    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char *IOSTContractStorage::Get(const char *key) {
    size_t gasUsed = 0;
    char *ret = CGet(sbxPtr, key, &gasUsed);
    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

int IOSTContractStorage::Del(const char *key) {
    size_t gasUsed = 0;
    int ret = CDel(sbxPtr, key, &gasUsed);
    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char *IOSTContractStorage::GlobalGet(const char *contract, const char *key) {
    size_t gasUsed = 0;
    char *ret = CGGet(sbxPtr, contract, key, &gasUsed);
    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

void NewIOSTContractStorage(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Context> context = isolate->GetCurrentContext();
    Local<Object> global = context->Global();

    Local<Value> val = global->GetInternalField(0);
    if (!val->IsExternal()) {
           std::cout << "NewIOSTContractStorage val error" << std::endl;
        return;
    }
    SandboxPtr sbx = static_cast<SandboxPtr>(Local<External>::Cast(val)->Value());

    IOSTContractStorage *ics = new IOSTContractStorage(sbx);

    Local<Object> self = args.Holder();
    self->SetInternalField(0, External::New(isolate, ics));

    args.GetReturnValue().Set(self);
}

void IOSTContractStorage_Put(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 2) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_Put invalid argument length.")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> key = args[0];
    if (!key->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_Put key must be string.")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> val = args[1];
    if (!val->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_Put value must be string.")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value keyStr(key);
    String::Utf8Value valStr(val);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTContractStorage_Put val error" << std::endl;
        return;
    }

    IOSTContractStorage *ics = static_cast<IOSTContractStorage *>(extVal->Value());
    int ret = ics->Put(*keyStr, *valStr);
    args.GetReturnValue().Set(ret);
}

void IOSTContractStorage_Get(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 1) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_Get invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> key = args[0];
    if (!key->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_Get key must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value keyStr(key);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTContractStorage_Get val error" << std::endl;
        return;
    }

    IOSTContractStorage *ics = static_cast<IOSTContractStorage *>(extVal->Value());
    char *val = ics->Get(*keyStr);
    if (val == nullptr) {
        args.GetReturnValue().SetNull();
    } else {
        args.GetReturnValue().Set(String::NewFromUtf8(isolate, val));
        free(val);
    }
}

void IOSTContractStorage_Del(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 1) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_Del invalid argument length.")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> key = args[0];
    if (!key->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_Del key must be string.")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value keyStr(key);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTContractStorage_Del val error" << std::endl;
        return;
    }

    IOSTContractStorage *ics = static_cast<IOSTContractStorage *>(extVal->Value());
    int ret = ics->Del(*keyStr);
    args.GetReturnValue().Set(ret);
}

void IOSTContractStorage_GGet(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 2) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_GGet invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> contract = args[0];
    if (!contract->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_GGet contract must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> key = args[1];
    if (!key->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_GGet key must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value contractStr(contract);
    String::Utf8Value keyStr(key);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTContractStorage_Get val error" << std::endl;
        return;
    }

    IOSTContractStorage *ics = static_cast<IOSTContractStorage *>(extVal->Value());
    char *val = ics->Get(*keyStr);
    if (val == nullptr) {
        args.GetReturnValue().SetNull();
    } else {
        args.GetReturnValue().Set(String::NewFromUtf8(isolate, val));
        free(val);
    }
}

void InitStorage(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    Local<FunctionTemplate> storageClass =
        FunctionTemplate::New(isolate, NewIOSTContractStorage);
    Local<String> storageClassName = String::NewFromUtf8(isolate, "IOSTStorage");
    storageClass->SetClassName(storageClassName);

    Local<ObjectTemplate> storageTpl = storageClass->InstanceTemplate();
    storageTpl->SetInternalFieldCount(1);
    storageTpl->Set(
        String::NewFromUtf8(isolate, "put"),
        FunctionTemplate::New(isolate, IOSTContractStorage_Put)
    );
    storageTpl->Set(
        String::NewFromUtf8(isolate, "get"),
        FunctionTemplate::New(isolate, IOSTContractStorage_Get)
    );
    storageTpl->Set(
            String::NewFromUtf8(isolate, "del"),
            FunctionTemplate::New(isolate, IOSTContractStorage_Del)
    );
    storageTpl->Set(
        String::NewFromUtf8(isolate, "globalGet"),
        FunctionTemplate::New(isolate, IOSTContractStorage_GGet)
    );


    globalTpl->Set(storageClassName, storageClass);
}