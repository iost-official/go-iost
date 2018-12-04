#ifndef IOST_V8_ENGINE_H
#define IOST_V8_ENGINE_H

#include <stddef.h>
#include <stdbool.h>
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
    void* isolate;
    void* allocator;
} IsolateWrapper;

typedef struct {
    const char *Value;
    const char *Err;
    bool isJson;
    size_t gasUsed;
} ValueTuple;

extern void init();
extern IsolateWrapperPtr newIsolate(CustomStartupData);
extern void releaseIsolate(IsolateWrapperPtr ptr);

// free memory
extern void lowMemoryNotification(IsolateWrapperPtr ptr);

extern SandboxPtr newSandbox(IsolateWrapperPtr ptr);
extern void loadVM(SandboxPtr ptr, int vmType);
extern void releaseSandbox(SandboxPtr ptr);

extern ValueTuple Execute(SandboxPtr ptr, const char *code, long long int expireTime);
extern void setJSPath(SandboxPtr ptr, const char *jsPath);
extern void setSandboxGasLimit(SandboxPtr ptr, size_t gasLimit);
extern void setSandboxMemLimit(SandboxPtr ptr, size_t memLimit);

// log
typedef char* (*consoleFunc)(SandboxPtr, const char *, const char *);
void InitGoConsole(consoleFunc);

// require
typedef char *(*requireFunc)(SandboxPtr, const char *);
void InitGoRequire(requireFunc);

// blockchain
typedef char* (*blockInfoFunc)(SandboxPtr, char **, size_t *);
typedef char* (*txInfoFunc)(SandboxPtr, char **, size_t *);
typedef char* (*contextInfoFunc)(SandboxPtr, char **, size_t *);
typedef char* (*callFunc)(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
typedef char* (*callWithAuthFunc)(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
typedef char* (*requireAuthFunc)(SandboxPtr, const char *, const char *, bool *, size_t *);
typedef char* (*receiptFunc)(SandboxPtr, const char *, size_t *);
typedef char* (*eventFunc)(SandboxPtr, const char *, size_t *);

void InitGoBlockchain(blockInfoFunc, txInfoFunc, contextInfoFunc, callFunc, callWithAuthFunc, requireAuthFunc, receiptFunc, eventFunc);

// storage
typedef char* (*putFunc)(SandboxPtr, const char *, const char *, const char *, size_t *);
typedef char* (*hasFunc)(SandboxPtr, const char *, const char *, bool *, size_t *);
typedef char* (*getFunc)(SandboxPtr, const char *, const char *, char **, size_t *);
typedef char* (*delFunc)(SandboxPtr, const char *, const char *, size_t *);
typedef char* (*mapPutFunc)(SandboxPtr, const char *, const char *, const char *, const char *, size_t *);
typedef char* (*mapHasFunc)(SandboxPtr, const char *, const char *, const char *, bool *, size_t *);
typedef char* (*mapGetFunc)(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
typedef char* (*mapDelFunc)(SandboxPtr, const char *, const char *, const char *, size_t *);
typedef char* (*mapKeysFunc)(SandboxPtr, const char *, const char *, char **, size_t *);
typedef char* (*mapLenFunc)(SandboxPtr, const char *, const char *, size_t *, size_t *);
typedef char* (*globalHasFunc)(SandboxPtr, const char *, const char *, const char *, bool *, size_t *);
typedef char* (*globalGetFunc)(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
typedef char* (*globalMapHasFunc)(SandboxPtr, const char *, const char *, const char *, const char *, bool *, size_t *);
typedef char* (*globalMapGetFunc)(SandboxPtr, const char *, const char *, const char *, const char *, char **, size_t *);
typedef char* (*globalMapKeysFunc)(SandboxPtr, const char *,  const char *, const char *, char **, size_t *);
typedef char* (*globalMapLenFunc)(SandboxPtr, const char *, const char *, const char *, size_t *, size_t *);

void InitGoStorage(putFunc, hasFunc, getFunc, delFunc,
    mapPutFunc, mapHasFunc, mapGetFunc, mapDelFunc, mapKeysFunc, mapLenFunc,
    globalHasFunc, globalGetFunc, globalMapHasFunc, globalMapGetFunc, globalMapKeysFunc, globalMapLenFunc);

// crypto
typedef char* (*sha3Func)(SandboxPtr, const char *, size_t *);

void InitGoCrypto(sha3Func);

extern int compile(SandboxPtr, const char *code, const char **compiledCode);
extern CustomStartupData createStartupData();
extern CustomStartupData createCompileStartupData();

#ifdef __cplusplus
}
#endif // __cplusplus

#endif // IOST_V8_ENGINE_H
