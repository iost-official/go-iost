#ifndef IOST_V8_ENGINE_H
#define IOST_V8_ENGINE_H

#include <stddef.h>
#include <stdbool.h>
#include <stdint.h>
#include "error.h"

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

typedef void* IsolateWrapperPtr;
typedef void* SandboxPtr;

typedef struct {
    const char *data;
    int raw_size;
} CustomStartupData;

typedef struct {
    char *data;
    int size;
} CStr;

typedef struct {
    void* isolate;
    void* allocator;
} IsolateWrapper;

typedef struct {
    CStr Value;
    CStr Err;
    bool isJson;
    size_t gasUsed;
} ValueTuple;

extern void init();
extern IsolateWrapperPtr newIsolate(CustomStartupData);
extern void releaseIsolate(IsolateWrapperPtr ptr);

// free memory
extern void lowMemoryNotification(IsolateWrapperPtr ptr);

extern SandboxPtr newSandbox(IsolateWrapperPtr ptr, int64_t);
extern void loadVM(SandboxPtr ptr, int vmType);
extern void releaseSandbox(SandboxPtr ptr);

extern ValueTuple Execute(SandboxPtr ptr, const CStr code, long long int expireTime);
extern void setJSPath(SandboxPtr ptr, const char *jsPath);
extern void setSandboxGasLimit(SandboxPtr ptr, size_t gasLimit);
extern void setSandboxMemLimit(SandboxPtr ptr, size_t memLimit);

// log
typedef char* (*consoleFunc)(SandboxPtr, const CStr, const CStr);
void InitGoConsole(consoleFunc);

// require
typedef char *(*requireFunc)(SandboxPtr, const CStr);
void InitGoRequire(requireFunc);

// blockchain
typedef char* (*blockInfoFunc)(SandboxPtr, CStr *, size_t *);
typedef char* (*txInfoFunc)(SandboxPtr, CStr *, size_t *);
typedef char* (*contextInfoFunc)(SandboxPtr, CStr *, size_t *);
typedef char* (*callFunc)(SandboxPtr, const CStr, const CStr, const CStr, CStr *, size_t *);
typedef char* (*callWithAuthFunc)(SandboxPtr, const CStr, const CStr, const CStr, CStr *, size_t *);
typedef char* (*requireAuthFunc)(SandboxPtr, const CStr, const CStr, bool *, size_t *);
typedef char* (*receiptFunc)(SandboxPtr, const CStr, size_t *);
typedef char* (*eventFunc)(SandboxPtr, const CStr, size_t *);

void InitGoBlockchain(blockInfoFunc, txInfoFunc, contextInfoFunc, callFunc, callWithAuthFunc, requireAuthFunc, receiptFunc, eventFunc);

// storage
typedef char* (*putFunc)(SandboxPtr, const CStr, const CStr, const CStr, size_t *);
typedef char* (*hasFunc)(SandboxPtr, const CStr, const CStr, bool *, size_t *);
typedef char* (*getFunc)(SandboxPtr, const CStr, const CStr, CStr *, size_t *);
typedef char* (*delFunc)(SandboxPtr, const CStr, const CStr, size_t *);
typedef char* (*mapPutFunc)(SandboxPtr, const CStr, const CStr, const CStr, const CStr, size_t *);
typedef char* (*mapHasFunc)(SandboxPtr, const CStr, const CStr, const CStr, bool *, size_t *);
typedef char* (*mapGetFunc)(SandboxPtr, const CStr, const CStr, const CStr, CStr *, size_t *);
typedef char* (*mapDelFunc)(SandboxPtr, const CStr, const CStr, const CStr, size_t *);
typedef char* (*mapKeysFunc)(SandboxPtr, const CStr, const CStr, CStr *, size_t *);
typedef char* (*mapLenFunc)(SandboxPtr, const CStr, const CStr, size_t *, size_t *);
typedef char* (*globalHasFunc)(SandboxPtr, const CStr, const CStr, const CStr, bool *, size_t *);
typedef char* (*globalGetFunc)(SandboxPtr, const CStr, const CStr, const CStr, CStr *, size_t *);
typedef char* (*globalMapHasFunc)(SandboxPtr, const CStr, const CStr, const CStr, const CStr, bool *, size_t *);
typedef char* (*globalMapGetFunc)(SandboxPtr, const CStr, const CStr, const CStr, const CStr, CStr *, size_t *);
typedef char* (*globalMapKeysFunc)(SandboxPtr, const CStr,  const CStr, const CStr, CStr *, size_t *);
typedef char* (*globalMapLenFunc)(SandboxPtr, const CStr, const CStr, const CStr, size_t *, size_t *);

void InitGoStorage(putFunc, hasFunc, getFunc, delFunc,
    mapPutFunc, mapHasFunc, mapGetFunc, mapDelFunc, mapKeysFunc, mapLenFunc,
    globalHasFunc, globalGetFunc, globalMapHasFunc, globalMapGetFunc, globalMapKeysFunc, globalMapLenFunc);

// crypto
typedef CStr (*sha3Func)(SandboxPtr, const CStr, size_t *);
typedef CStr (*sha3HexFunc)(SandboxPtr, const CStr, size_t *);
typedef CStr (*ripemd160HexFunc)(SandboxPtr, const CStr, size_t *);
typedef int (*verifyFunc)(SandboxPtr, const CStr, const CStr, const CStr, const CStr, size_t *);

void InitGoCrypto(sha3Func, sha3HexFunc, ripemd160HexFunc, verifyFunc);

extern int compile(SandboxPtr, const CStr code, CStr *compiledCode, CStr *errMsg);
extern int validate(SandboxPtr ptr, const CStr code, const CStr abi, CStr *result, CStr *errMsg);
extern CustomStartupData createStartupData();
extern CustomStartupData createCompileStartupData();

#ifdef __cplusplus
}
#endif // __cplusplus

#endif // IOST_V8_ENGINE_H
