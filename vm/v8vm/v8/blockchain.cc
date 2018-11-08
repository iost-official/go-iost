#include "blockchain.h"
#include <iostream>

static blockInfoFunc CBlkInfo = nullptr;
static txInfoFunc CTxInfo = nullptr;
static contextInfoFunc CCtxInfo = nullptr;
static callFunc CCall = nullptr;
static callWithAuthFunc CCallWA = nullptr;
static requireAuthFunc CRequireAuth = nullptr;
static receiptFunc CReceipt = nullptr;
static eventFunc CEvent = nullptr;

void InitGoBlockchain(blockInfoFunc blkInfo, txInfoFunc txInfo, contextInfoFunc contextInfo,
		callFunc call, CallWithAuthFunc callWA,
        requireAuthFunc requireAuth, receiptFunc receipt, eventFunc event) {
    CBlkInfo = blkInfo;
    CTxInfo = txInfo;
    CCtxInfo = contextInfo;
    CCall = call;
    CCallWA = callWA;
    CRequireAuth = requireAuth;
	CReceipt = receipt;
	CEvent = event;
}

char* IOSTBlockchain::BlockInfo(char **result) {
    size_t gasUsed = 0;

    char* ret = CBlkInfo(sbxPtr, result, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char* IOSTBlockchain::TxInfo(char **result) {
    size_t gasUsed = 0;

    char* ret = CTxInfo(sbxPtr, result, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char* IOSTBlockchain::ContextInfo(char **result) {
    size_t gasUsed = 0;

    char* ret = CCtxInfo(sbxPtr, result, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char* IOSTBlockchain::Call(const char *contract, const char *api, const char *args, char **result) {
    size_t gasUsed = 0;
    char* ret = CCall(sbxPtr, contract, api, args, result, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char* IOSTBlockchain::CallWithAuth(const char *contract, const char *abi, const char *args, char **result) {
    size_t gasUsed = 0;
    char* ret = CCallWA(sbxPtr, contract, api, args, result, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char* IOSTBlockchain::RequireAuth(const char *accountID, const char *permission, bool *result) {
    size_t gasUsed = 0;
    char* ret = CRequireAuth(sbxPtr, pubKey, permission, result, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char* IOSTBlockchain::Receipt(const char *content) {
    size_t gasUsed = 0;
    char* ret = CReceipt(sbxPtr, content, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char* IOSTBlockchain::Event(const char *content) {
    size_t gasUsed = 0;
    char* ret = CEvent(sbxPtr, content, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

void NewIOSTBlockchain(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Context> context = isolate->GetCurrentContext();
    Local<Object> global = context->Global();

    Local<Value> val = global->GetInternalField(0);
    if (!val->IsExternal()) {
           std::cout << "NewIOSTBlockchain val error" << std::endl;
        return;
    }
    SandboxPtr sbx = static_cast<SandboxPtr>(Local<External>::Cast(val)->Value());

    IOSTBlockchain *bc = new IOSTBlockchain(sbx);

    Local<Object> self = args.Holder();
    self->SetInternalField(0, External::New(isolate, bc));

    args.GetReturnValue().Set(self);
}

void IOSTBlockchain_blockInfo(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    String::Utf8Value resultStr("");

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_blockInfo val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());

    char *ret = bc->BlockInfo(resultStr);
    if (ret != nullptr) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, ret)
        );
        isolate->ThrowException(err);
        return;
    }
    args.GetReturnValue().Set(String::NewFromUtf8(isolate, *resultStr));
}

void IOSTBlockchain_txInfo(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    String::Utf8Value resultStr("");

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_txInfo val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->TxInfo(resultStr);
    if (ret != nullptr) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, ret)
        );
        isolate->ThrowException(err);
        return;
    }
    args.GetReturnValue().Set(String::NewFromUtf8(isolate, *resultStr));
}

void IOSTBlockchain_contextInfo(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    String::Utf8Value resultStr("");

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_contextInfo val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->ContextInfo(resultStr);
    if (ret != nullptr) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, ret)
        );
        isolate->ThrowException(err);
        return;
    }
    args.GetReturnValue().Set(String::NewFromUtf8(isolate, *resultStr));
}

void IOSTBlockchain_call(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 3) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_call invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> contract = args[0];
    if (!contract->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_call contract must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> api = args[1];
    if (!api->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_call api must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> arg = args[2];
    if (!arg->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_call arg must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value contractStr(contract);
    String::Utf8Value apiStr(api);
    String::Utf8Value argStr(arg);
    String::Utf8Value resultStr("");

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_call val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->Call(*contractStr, *apiStr, *argStr, resultStr);
    if (ret != nullptr) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, ret)
        );
        isolate->ThrowException(err);
        return;
    }
    args.GetReturnValue().Set(String::NewFromUtf8(isolate, *resultStr));
}


void IOSTBlockchain_callWithAuth(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 3) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_callWithAuth invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> contract = args[0];
    if (!contract->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_callWithAuth contract must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> api = args[1];
    if (!api->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_callWithAuth api must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> arg = args[2];
    if (!arg->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_callWithAuth arg must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value contractStr(contract);
    String::Utf8Value apiStr(api);
    String::Utf8Value argStr(arg);
    String::Utf8Value resultStr("");

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_callWithAuth val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->CallWithAuth(*contractStr, *apiStr, *argStr, resultStr);
    if (ret != nullptr) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, ret)
        );
        isolate->ThrowException(err);
        return;
    }
    args.GetReturnValue().Set(String::NewFromUtf8(isolate, *resultStr));
}

void IOSTBlockchain_requireAuth(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 2) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_requireAuth invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> accountID = args[0];
    if (!accountID->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_requireAuth accountID must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> permission = args[1];
    if (!permission->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_requireAuth permission must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value accountIDStr(accountID);
    String::Utf8Value permissionStr(permission);
    bool result;

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_requireAuth val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->RequireAuth(*accountIDStr, *permissionStr, &result);
    if (ret != nullptr) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, ret)
        );
        isolate->ThrowException(err);
        return;
    }
    args.GetReturnValue().Set(result);
}

void IOSTBlockchain_receipt(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 1) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_receipt invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> content = args[1];
    if (!content->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_receipt content must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value contentStr(content);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_receipt val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->Receipt(*content);
    if (ret != nullptr) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, ret)
        );
        isolate->ThrowException(err);
        return;
    }
    args.GetReturnValue().SetNull();
}

void IOSTBlockchain_event(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 1) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_event invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> content = args[1];
    if (!content->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_event content must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value contentStr(content);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_event val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->Event(*content);
    if (ret != nullptr) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, ret)
        );
        isolate->ThrowException(err);
        return;
    }
    args.GetReturnValue().SetNull();
}

void InitBlockchain(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    Local<FunctionTemplate> blockchainClass =
        FunctionTemplate::New(isolate, NewIOSTBlockchain);
    Local<String> blockchainClassName = String::NewFromUtf8(isolate, "IOSTBlockchain");
    blockchainClass->SetClassName(blockchainClassName);

    Local<ObjectTemplate> blockchainTpl = blockchainClass->InstanceTemplate();
    blockchainTpl->SetInternalFieldCount(1);
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "blockInfo"),
        FunctionTemplate::New(isolate, IOSTBlockchain_blockInfo)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "txInfo"),
        FunctionTemplate::New(isolate, IOSTBlockchain_txInfo)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "contextInfo"),
        FunctionTemplate::New(isolate, IOSTBlockchain_contextInfo)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "call"),
        FunctionTemplate::New(isolate, IOSTBlockchain_call)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "callWithAuth"),
        FunctionTemplate::New(isolate, IOSTBlockchain_callWithAuth)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "requireAuth"),
        FunctionTemplate::New(isolate, IOSTBlockchain_requireAuth)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "receipt"),
        FunctionTemplate::New(isolate, IOSTBlockchain_receipt)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "event"),
        FunctionTemplate::New(isolate, IOSTBlockchain_event)
    );

    globalTpl->Set(blockchainClassName, blockchainClass);
}
