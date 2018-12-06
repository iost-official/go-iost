#ifndef IOST_V8_CRYPTO_H
#define IOST_V8_CRYPTO_H

#include "sandbox.h"

// This Class Provide Console.Log Function so JS code can use Go log.
void InitCrypto(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewCrypto(const FunctionCallbackInfo<Value> &info);

class IOSTCrypto {
private:
    SandboxPtr sbxPtr;
public:
    IOSTCrypto(SandboxPtr ptr): sbxPtr(ptr) {}

    char* sha3(const char *msg);
};

#endif // IOST_V8_CRYPTO_H