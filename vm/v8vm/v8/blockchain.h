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

    char* blockInfo(char **result);
	char* txInfo(char **result);
	char* contextInfo(char **result);
	char* call(const char *contract, const char *api, const char *args, char **result);
	char* callWithAuth(const char *contract, const char *api, const char *args, char **result);
	char* requireAuth(const char *accountID, const char *permission, bool *result);
	char* receipt(const char *content);
	char* event(const char *content);
};

#endif // IOST_V8_BLOCKCHAIN_H
