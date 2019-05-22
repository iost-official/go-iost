#ifndef IOST_V8_BLOCKCHAIN_H
#define IOST_V8_BLOCKCHAIN_H

#include "sandbox.h"
#include "stddef.h"

using namespace v8;

void InitBlockchain(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTBlockchain(const FunctionCallbackInfo<Value> &args);

// This Class wraps Go BlockChain function so JS contract can call them.
class IOSTBlockchain {
private:
    SandboxPtr sbxPtr;
public:
    IOSTBlockchain(SandboxPtr ptr): sbxPtr(ptr) {}

    char* BlockInfo(CStr *result);
    char* TxInfo(CStr *result);
    char* ContextInfo(CStr *result);
    char* Call(const CStr contract, const CStr api, const CStr args, CStr *result);
    char* CallWithAuth(const CStr contract, const CStr api, const CStr args, CStr *result);
    char* RequireAuth(const CStr accountID, const CStr permission, bool *result);
    char* Receipt(const CStr content);
    char* Event(const CStr content);
};

#endif // IOST_V8_BLOCKCHAIN_H
