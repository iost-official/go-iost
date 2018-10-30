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

extern SandboxPtr newSandbox(IsolateWrapperPtr ptr);
extern void loadVM(SandboxPtr ptr, int vmType);
extern void releaseSandbox(SandboxPtr ptr);

extern ValueTuple Execute(SandboxPtr ptr, const char *code, long long int expireTime);
extern void setJSPath(SandboxPtr ptr, const char *jsPath);
extern void setSandboxGasLimit(SandboxPtr ptr, size_t gasLimit);
extern void setSandboxMemLimit(SandboxPtr ptr, size_t memLimit);

// log
typedef int (*consoleFunc)(SandboxPtr, const char *, const char *);
void InitGoConsole(consoleFunc);

// require
typedef char *(*requireFunc)(SandboxPtr, const char *);
void InitGoRequire(requireFunc);

// blockchain
typedef int (*transferFunc)(SandboxPtr, const char *, const char *, const char *, size_t *);
typedef int (*withdrawFunc)(SandboxPtr, const char *, const char *, size_t *);
typedef int (*depositFunc)(SandboxPtr, const char *, const char *, size_t *);
typedef int (*topUpFunc)(SandboxPtr, const char *, const char *, const char *, size_t *);
typedef int (*countermandFunc)(SandboxPtr, const char *, const char *, const char *, size_t *);
typedef int (*blockInfoFunc)(SandboxPtr, char **, size_t *);
typedef int (*txInfoFunc)(SandboxPtr, char **, size_t *);
typedef int (*callFunc)(SandboxPtr, const char *, const char *, const char *, char **, size_t *);
typedef int (*requireAuthFunc)(SandboxPtr, const char *, const bool *, size_t *);
typedef int (*grantServiFunc)(SandboxPtr, const char *, const char *, size_t *);
void InitGoBlockchain(transferFunc, withdrawFunc,
                        depositFunc, topUpFunc, countermandFunc,
                        blockInfoFunc, txInfoFunc, callFunc, callFunc, requireAuthFunc, grantServiFunc);

// storage
typedef int (*putFunc)(SandboxPtr, const char *, const char *, size_t *);
typedef char *(*getFunc)(SandboxPtr, const char *, size_t *);
typedef int (*delFunc)(SandboxPtr, const char *, size_t *);
typedef int (*mapPutFunc)(SandboxPtr, const char *, const char *, const char *, size_t *);
typedef bool (*mapHasFunc)(SandboxPtr, const char *, const char *, size_t *);
typedef char *(*mapGetFunc)(SandboxPtr, const char *, const char *, size_t *);
typedef int (*mapDelFunc)(SandboxPtr, const char *, const char *, size_t *);
typedef char *(*mapKeysFunc)(SandboxPtr, const char *, size_t *);
typedef char *(*globalGetFunc)(SandboxPtr, const char *, const char *, size_t *);
void InitGoStorage(putFunc, getFunc, delFunc,
    mapPutFunc, mapHasFunc, mapGetFunc, mapDelFunc, mapKeysFunc,
    globalGetFunc);

extern void goMapLen(SandboxPtr, const char *, size_t *);
extern void goGlobalMapGet(SandboxPtr, const char *, const char *, const char *, size_t *);
extern void goGlobalMapKeys(SandboxPtr, const char *, const char *, size_t *);
extern void goGlobalMapLen(SandboxPtr, const char *, const char *, size_t *);

extern int compile(SandboxPtr, const char *code, const char **compiledCode);
extern CustomStartupData createStartupData();
extern CustomStartupData createCompileStartupData();

#ifdef __cplusplus
}
#endif // __cplusplus

#endif // IOST_V8_ENGINE_H
