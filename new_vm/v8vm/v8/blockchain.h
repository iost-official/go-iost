#ifndef IOST_V8_BLOCKCHAIN_H
#define IOST_V8_BLOCKCHAIN_H

#include "v8.h"
#include "vm.h"
#include "stddef.h"

using namespace v8;

void InitBlockChain(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTBlockchain(const FunctionCallbackInfo<Value> &args);

class IOSTBlockchain {
private:
    SandboxPtr sbx;
public:
    IOSTBlockchain(SandboxPtr ptr): sbx(ptr) {}

    int Transfer(char *from, char *to, char *amount);
    int Withdraw(char *to, char *amount);
    int Deposit(char *from, char *amount);
    int TopUp(char *contract, char *from, char *amount);
    int Countermand(char *contract, char *to, char *amount);
    char *BlockInfo();
    char *TxInfo();
    char *Call(char *contract, char *api, char *args);
};

#endif // IOST_V8_BLOCKCHAIN_H