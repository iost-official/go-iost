#include "storage.h"
#include <iostream>

static putFunc CPut = nullptr;
static getFunc CGet = nullptr;
static delFunc CDel = nullptr;

void InitGoStorage(putFunc put, getFunc get, delFunc del) {
    CPut = put;
    CGet = get;
    CDel = del;
}

int IOSTContractStorage::Put(char *key, char *value) {
    size_t gasCount = 0;
    int ret = CPut(sbx, key, value, &gasCount);
    return ret;
}

char *IOSTContractStorage::Get(char *key) {
    size_t gasCount = 0;
    char *ret = CGet(sbx, key, &gasCount);
    return ret;
}

int IOSTContractStorage::Del(char *key) {
    size_t gasCount = 0;
    int ret = CDel(sbx, key, &gasCount);
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
            String::NewFromUtf8(isolate, "IOSTContractStorage_Get invalid argument length.")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> key = args[0];
    if (!key->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTContractStorage_Get key must be string.")
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


    globalTpl->Set(storageClassName, storageClass);
}