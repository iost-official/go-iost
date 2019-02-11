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

    CStr sha3(const CStr msg);
    int verify(const CStr algo, const CStr msg, const CStr sig, const CStr pubkey);
};

#endif // IOST_V8_CRYPTO_H