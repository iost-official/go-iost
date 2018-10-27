#include "blockchain.h"
#include <iostream>

static transferFunc CTransfer = nullptr;
static withdrawFunc CWithdraw = nullptr;
static depositFunc CDeposit = nullptr;
static topUpFunc CTopUp = nullptr;
static countermandFunc CCountermand = nullptr;
static blockInfoFunc CBlkInfo = nullptr;
static txInfoFunc CTxInfo = nullptr;
static callFunc CCall = nullptr;
static callFunc CCallWR = nullptr;
static requireAuthFunc CRequireAuth = nullptr;
static grantServiFunc CGrantServi = nullptr;

void InitGoBlockchain(transferFunc transfer, withdrawFunc withdraw,
                        depositFunc deposit, topUpFunc topUp, countermandFunc countermand,
                        blockInfoFunc blkInfo, txInfoFunc txInfo, callFunc call, callFunc callWR,
                        requireAuthFunc requireAuth, grantServiFunc grantServi) {
    CTransfer = transfer;
    CWithdraw = withdraw;
    CDeposit = deposit;
    CTopUp = topUp;
    CCountermand = countermand;
    CBlkInfo = blkInfo;
    CTxInfo = txInfo;
    CCall = call;
    CCallWR = callWR;
    CRequireAuth = requireAuth;
    CGrantServi = grantServi;
}

int IOSTBlockchain::Transfer(const char *from, const char *to, const char *amount) {
    size_t gasUsed = 0;
    int ret = CTransfer(sbxPtr, from, to, amount, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

int IOSTBlockchain::Withdraw(const char *to, const char *amount) {
    size_t gasUsed = 0;
    int ret = CWithdraw(sbxPtr, to, amount, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

int IOSTBlockchain::Deposit(const char *from, const char *amount) {
    size_t gasUsed = 0;
    int ret = CDeposit(sbxPtr, from, amount, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

int IOSTBlockchain::TopUp(const char *contract, const char *from, const char *amount) {
    size_t gasUsed = 0;
    int ret = CTopUp(sbxPtr, contract, from, amount, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

int IOSTBlockchain::Countermand(const char *contract, const char *to, const char *amount) {
    size_t gasUsed = 0;
    int ret = CCountermand(sbxPtr, contract, to, amount, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ret;
}

char *IOSTBlockchain::BlockInfo() {
    size_t gasUsed = 0;
    char *info = nullptr;

    int ret = CBlkInfo(sbxPtr, &info, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return info;
}

char *IOSTBlockchain::TxInfo() {
    size_t gasUsed = 0;
    char *info = nullptr;
    int ret =
    CTxInfo(sbxPtr, &info, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return info;
}

char *IOSTBlockchain::Call(const char *contract, const char *api, const char *args) {
    size_t gasUsed = 0;
    char *result = nullptr;
    int ret = CCall(sbxPtr, contract, api, args, &result, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return result;
}

char *IOSTBlockchain::CallWithReceipt(const char *contract, const char *api, const char *args) {
    size_t gasUsed = 0;
    char *result = nullptr;
    int ret = CCallWR(sbxPtr, contract, api, args, &result, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return result;
}

bool IOSTBlockchain::RequireAuth(const char *pubKey, const char *permission) {
    size_t gasUsed = 0;
    bool ok = false;
    int ret = CRequireAuth(sbxPtr, pubKey, permission, &ok, &gasUsed);

    Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
    sbx->gasUsed += gasUsed;
    return ok;
}

int IOSTBlockchain::GrantServi(const char *pubKey, const char *amount) {
    size_t gasUsed = 0;
    int ret = CGrantServi(sbxPtr, pubKey, amount, &gasUsed);

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

void IOSTBlockchain_transfer(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 3) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_transfer invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> from = args[0];
    if (!from->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_transfer from must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> to = args[1];
    if (!to->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_transfer to must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> amount = args[2];
    if (!amount->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_transfer amount must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value fromStr(from);
    String::Utf8Value toStr(to);
    String::Utf8Value amountStr(amount);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_transfer val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    int ret = bc->Transfer(*fromStr, *toStr, *amountStr);
    args.GetReturnValue().Set(ret);
}

void IOSTBlockchain_withdraw(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 2) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_withdraw invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> to = args[0];
    if (!to->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_withdraw to must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> amount = args[1];
    if (!amount->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_withdraw amount must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value toStr(to);
    String::Utf8Value amountStr(amount);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_withdraw val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    int ret = bc->Withdraw(*toStr, *amountStr);
    args.GetReturnValue().Set(ret);
}

void IOSTBlockchain_deposit(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 2) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_deposit invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> from = args[0];
    if (!from->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_deposit from must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> amount = args[1];
    if (!amount->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_deposit amount must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value fromStr(from);
    String::Utf8Value amountStr(amount);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_deposit val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    int ret = bc->Deposit(*fromStr, *amountStr);
    args.GetReturnValue().Set(ret);
}

void IOSTBlockchain_topUp(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 3) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_topUp invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> contract = args[0];
    if (!contract->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_topUp from must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> from = args[1];
    if (!from->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_topUp to must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> amount = args[2];
    if (!amount->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_topUp amount must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value contractStr(contract);
    String::Utf8Value fromStr(from);
    String::Utf8Value amountStr(amount);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_topUp val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    int ret = bc->TopUp(*contractStr, *fromStr, *amountStr);
    args.GetReturnValue().Set(ret);
}

void IOSTBlockchain_countermand(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 3) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_countermand invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> contract = args[0];
    if (!contract->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_countermand from must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> to = args[1];
    if (!to->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_countermand to must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> amount = args[2];
    if (!amount->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_countermand amount must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value contractStr(contract);
    String::Utf8Value toStr(to);
    String::Utf8Value amountStr(amount);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_countermand val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    int ret = bc->Countermand(*contractStr, *toStr, *amountStr);
    args.GetReturnValue().Set(ret);
}

void IOSTBlockchain_blockInfo(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_blockInfo val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *blkInfo = bc->BlockInfo();
    if (blkInfo == nullptr) {
        args.GetReturnValue().SetNull();
    } else {
        args.GetReturnValue().Set(String::NewFromUtf8(isolate, blkInfo));
        free(blkInfo);
    }
}

void IOSTBlockchain_txInfo(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_txInfo val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *txInfo = bc->TxInfo();
    if (txInfo == nullptr) {
        args.GetReturnValue().SetNull();
    } else {
        args.GetReturnValue().Set(String::NewFromUtf8(isolate, txInfo));
        free(txInfo);
    }
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

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_call val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->Call(*contractStr, *apiStr, *argStr);
    if (ret == nullptr) {
        args.GetReturnValue().SetNull();
    } else {
        args.GetReturnValue().Set(String::NewFromUtf8(isolate, ret));
        free(ret);
    }
}

//todo refine this code
void IOSTBlockchain_callWithReceipt(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 3) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_callWithReceipt invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> contract = args[0];
    if (!contract->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_callWithReceipt contract must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> api = args[1];
    if (!api->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_callWithReceipt api must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> arg = args[2];
    if (!arg->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_callWithReceipt arg must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value contractStr(contract);
    String::Utf8Value apiStr(api);
    String::Utf8Value argStr(arg);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_callWithReceipt val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    char *ret = bc->CallWithReceipt(*contractStr, *apiStr, *argStr);
    if (ret == nullptr) {
        args.GetReturnValue().SetNull();
    } else {
        args.GetReturnValue().Set(String::NewFromUtf8(isolate, ret));
        free(ret);
    }
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

    Local<Value> pubKey = args[0];
    if (!pubKey->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_requireAuth pubKey must be string")
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

    String::Utf8Value pubKeyStr(pubKey);
    String::Utf8Value permissionStr(permission);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_requireAuth val error" << std::endl;
        return;
    }

    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    bool ret = bc->RequireAuth(*pubKeyStr, *permissionStr);

    args.GetReturnValue().Set(ret);
}

void IOSTBlockchain_grantServi(const FunctionCallbackInfo<Value> &args) {
    Isolate *isolate = args.GetIsolate();
    Local<Object> self = args.Holder();

    if (args.Length() != 2) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_grantServi invalid argument length")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> pubKey = args[0];
    if (!pubKey->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_grantServi pubKey must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    Local<Value> amount = args[1];
    if (!amount->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_grantServi amount must be string")
        );
        isolate->ThrowException(err);
        return;
    }

    String::Utf8Value pubKeyStr(pubKey);
    String::Utf8Value amountStr(amount);

    Local<External> extVal = Local<External>::Cast(self->GetInternalField(0));
    if (!extVal->IsExternal()) {
        std::cout << "IOSTBlockchain_grantServi val error" << std::endl;
        return;
    }


    IOSTBlockchain *bc = static_cast<IOSTBlockchain *>(extVal->Value());
    int ret = bc->GrantServi(*pubKeyStr, *amountStr);

    args.GetReturnValue().Set(ret);
}

void InitBlockchain(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
    Local<FunctionTemplate> blockchainClass =
        FunctionTemplate::New(isolate, NewIOSTBlockchain);
    Local<String> blockchainClassName = String::NewFromUtf8(isolate, "IOSTBlockchain");
    blockchainClass->SetClassName(blockchainClassName);

    Local<ObjectTemplate> blockchainTpl = blockchainClass->InstanceTemplate();
    blockchainTpl->SetInternalFieldCount(1);
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "transfer"),
        FunctionTemplate::New(isolate, IOSTBlockchain_transfer)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "withdraw"),
        FunctionTemplate::New(isolate, IOSTBlockchain_withdraw)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "deposit"),
        FunctionTemplate::New(isolate, IOSTBlockchain_deposit)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "topUp"),
        FunctionTemplate::New(isolate, IOSTBlockchain_topUp)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "countermand"),
        FunctionTemplate::New(isolate, IOSTBlockchain_countermand)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "blockInfo"),
        FunctionTemplate::New(isolate, IOSTBlockchain_blockInfo)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "txInfo"),
        FunctionTemplate::New(isolate, IOSTBlockchain_txInfo)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "call"),
        FunctionTemplate::New(isolate, IOSTBlockchain_call)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "callWithReceipt"),
        FunctionTemplate::New(isolate, IOSTBlockchain_callWithReceipt)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "requireAuth"),
        FunctionTemplate::New(isolate, IOSTBlockchain_requireAuth)
    );
    blockchainTpl->Set(
        String::NewFromUtf8(isolate, "grantServi"),
        FunctionTemplate::New(isolate, IOSTBlockchain_grantServi)
    );


    globalTpl->Set(blockchainClassName, blockchainClass);
}
