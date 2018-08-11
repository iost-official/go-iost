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

void InitGoBlockchain(transferFunc transfer, withdrawFunc withdraw,
                        depositFunc deposit, topUpFunc topUp, countermandFunc countermand,
                        blockInfoFunc blkInfo, txInfoFunc txInfo, callFunc call) {
    CTransfer = transfer;
    CWithdraw = withdraw;
    CDeposit = deposit;
    CTopUp = topUp;
    CCountermand = countermand;
    CBlkInfo = blkInfo;
    CTxInfo = txInfo;
    CCall = call;
}

int IOSTBlockchain::Transfer(char *from, char *to, char *amount) {
    size_t gasCount = 0;
    int ret = CTransfer(sbx, from, to, amount, &gasCount);
    return ret;
}

int IOSTBlockchain::Withdraw(char *to, char *amount) {
    size_t gasCount = 0;
    int ret = CWithdraw(sbx, to, amount, &gasCount);
    return ret;
}

int IOSTBlockchain::Deposit(char *from, char *amount) {
    size_t gasCount = 0;
    int ret = CDeposit(sbx, from, amount, &gasCount);
    return ret;
}

int IOSTBlockchain::TopUp(char *contract, char *from, char *amount) {
    size_t gasCount = 0;
    int ret = CTopUp(sbx, contract, from, amount, &gasCount);
    return ret;
}

int IOSTBlockchain::Countermand(char *contract, char *to, char *amount) {
    size_t gasCount = 0;
    int ret = CCountermand(sbx, contract, to, amount, &gasCount);
    return ret;
}

char *IOSTBlockchain::BlockInfo() {
    size_t gasCount = 0;
    char *blkInfo = CBlkInfo(sbx, &gasCount);
    return blkInfo;
}

char *IOSTBlockchain::TxInfo() {
    size_t gasCount = 0;
    char *txInfo = CTxInfo(sbx, &gasCount);
    return txInfo;
}

char *IOSTBlockchain::Call(char *contract, char *api, char *args) {
    size_t gasCount = 0;
    char *result = CCall(sbx, contract, api, args, &gasCount);
    return result;
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
            String::NewFromUtf8(isolate, "IOSTBlockchain_call from must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> api = args[1];
    if (!api->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_call to must be string")
        );
        isolate->ThrowException(err);
        return;
    }
    Local<Value> arg = args[2];
    if (!arg->IsString()) {
        Local<Value> err = Exception::Error(
            String::NewFromUtf8(isolate, "IOSTBlockchain_call amount must be string")
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

    globalTpl->Set(blockchainClassName, blockchainClass);
}
