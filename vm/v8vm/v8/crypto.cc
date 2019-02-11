#include "crypto.h"
#include <iostream>

static sha3Func CSha3 = nullptr;
static verifyFunc CVerify = nullptr;

void InitGoCrypto(sha3Func sha3, verifyFunc verify) {
    CSha3 = sha3;
    CVerify = verify;
}

CStr IOSTCrypto::sha3(const CStr msg) {
    size_t gasUsed;
    CStr ret = CSha3(sbxPtr, msg, &gasUsed);
    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

int IOSTCrypto::verify(const CStr algo, const CStr msg, const CStr sig, const CStr pubkey) {
    size_t gasUsed;
    int ret = CVerify(sbxPtr, algo, msg, sig, pubkey, &gasUsed);
    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

void NewCrypto(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Context> context = isolate->GetCurrentContext();
    Local<Object> global = context->Global();

    Local<Value> val = global->GetInternalField(0);
    if (!val->IsExternal()) {
        std::cout << "NewCrypto val error" << std::endl;
        return;
    }
    SandboxPtr sbx = static_cast<SandboxPtr>(Local<External>::Cast(val)->Value());

    IOSTCrypto *ic = new IOSTCrypto(sbx);

    Local<Object> self = args.Holder();
    self->SetInternalField(0, External::New(isolate, ic));

    args.GetReturnValue().Set(self);
}

void IOSTCrypto_sha3(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 1) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTCrypto_sha3 invalid argument length.")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> msg = args[0];
    if (!msg->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTCrypto_sha3 msg must be string.")
        );
        isolate->ThrowException(err);
        return;
    }

    NewCStrChecked(msgStr, msg, isolate);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTCrypto_sha3 val error" << std::endl;
        return;
    }

    IOSTCrypto *ic = static_cast<IOSTCrypto *>(extVal->Value());
    CStr ret = ic->sha3(msgStr);
    if (ret.data != nullptr) {
        args.GetReturnValue().Set(String::NewFromUtf8(isolate, ret.data, String::kNormalString, ret.size));
        free(ret.data);
        return;
    }
    args.GetReturnValue().SetNull();
}

void IOSTCrypto_verify(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 4) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTCrypto_verify invalid argument length.")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> algo = args[0];
    if (!algo->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTCrypto_verify algo must be string.")
        );
        isolate->ThrowException(err);
        return;
    }
    NewCStrChecked(algoStr, algo, isolate);

    Local<Value> msg = args[1];
    if (!msg->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTCrypto_verify msg must be string.")
        );
        isolate->ThrowException(err);
        return;
    }
    NewCStrChecked(msgStr, msg, isolate);

    Local<Value> sig = args[2];
    if (!sig->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTCrypto_verify sig must be string.")
        );
        isolate->ThrowException(err);
        return;
    }
    NewCStrChecked(sigStr, sig, isolate);

    Local<Value> pubkey = args[3];
    if (!pubkey->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTCrypto_verify pubkey must be string.")
        );
        isolate->ThrowException(err);
        return;
    }
    NewCStrChecked(pubkeyStr, pubkey, isolate);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTCrypto_verify val error" << std::endl;
        return;
    }

    IOSTCrypto *ic = static_cast<IOSTCrypto *>(extVal->Value());
    int ret = ic->verify(algoStr, msgStr, sigStr, pubkeyStr);
    args.GetReturnValue().Set(ret);
}

void InitCrypto(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    Local<FunctionTemplate> cryptoClass =
        FunctionTemplate::New(isolate, NewCrypto);
    Local<String> cryptoClassName = String::NewFromUtf8(isolate, "_IOSTCrypto");
    cryptoClass->SetClassName(cryptoClassName);

    Local<ObjectTemplate> cryptoTpl = cryptoClass->InstanceTemplate();
    cryptoTpl->SetInternalFieldCount(2);
    cryptoTpl->Set(
        String::NewFromUtf8(isolate, "sha3"),
        FunctionTemplate::New(isolate, IOSTCrypto_sha3)
    );
    cryptoTpl->Set(
        String::NewFromUtf8(isolate, "verify"),
        FunctionTemplate::New(isolate, IOSTCrypto_verify)
    );

    globalTpl->Set(cryptoClassName, cryptoClass);
}