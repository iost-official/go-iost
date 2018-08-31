#ifndef IOST_V8_BLOCKCHAIN_H
#define IOST_V8_BLOCKCHAIN_H

#include "sandbox.h"
#include "stddef.h"

using namespace v8;

void InitBlockchain(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTBlockchain(const FunctionCallbackInfo<Value> &args);

class IOSTBlockchain {
private:
    SandboxPtr sbxPtr;
public:
    IOSTBlockchain(SandboxPtr ptr): sbxPtr(ptr) {}

    int Transfer(const char *, const char *, const char *);
    int Withdraw(const char *, const char *);
    int Deposit(const char *, const char *);
    int TopUp(const char *, const char *, const char *);
    int Countermand(const char *, const char *, const char *);
    char *BlockInfo();
    char *TxInfo();
    char *Call(const char *, const char *, const char *);
    char *CallWithReceipt(const char *, const char *, const char *);
    bool RequireAuth(const char *pubKey);
};

#endif // IOST_V8_BLOCKCHAIN_H