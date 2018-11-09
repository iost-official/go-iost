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

    char* BlockInfo(char **result);
	char* TxInfo(char **result);
	char* ContextInfo(char **result);
	char* Call(const char *contract, const char *api, const char *args, char **result);
	char* CallWithAuth(const char *contract, const char *api, const char *args, char **result);
	char* RequireAuth(const char *accountID, const char *permission, bool *result);
	char* Receipt(const char *content);
	char* Event(const char *content);
};

#endif // IOST_V8_BLOCKCHAIN_H
