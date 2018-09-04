#ifndef IOST_V8_ENGINE_H
#define IOST_V8_ENGINE_H

#include <stddef.h>
#include <stdbool.h>
#include "error.h"

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

typedef void* IsolatePtr;
typedef void* SandboxPtr;

typedef struct {
    const char *Value;
    const char *Err;
    bool isJson;
    size_t gasUsed;
} ValueTuple;

typedef enum {
    kUndefined = 0,
    kNull,
    kName,
    kString,
    kSymbol,
    kFunction,
    kArray,
    kObject,
    kBoolean,
    kNumber,
    kExternal,
    kInt32,
    kUint32,
    kDate,
    kArgumentsObject,
    kBooleanObject,
    kNumberObject,
    kStringObject,
    kSymbolObject,
    kNativeError,
    kRegExp,
    kAsyncFunction,
    kGeneratorFunction,
    kGeneratorObject,
    kPromise,
    kMap,
    kSet,
    kMapIterator,
    kSetIterator,
    kWeakMap,
    kWeakSet,
    kArrayBuffer,
    kArrayBufferView,
    kTypedArray,
    kUint8Array,
    kUint8ClampedArray,
    kInt8Array,
    kUint16Array,
    kInt16Array,
    kUint32Array,
    kInt32Array,
    kFloat32Array,
    kFloat64Array,
    kDataView,
    kSharedArrayBuffer,
    kProxy,
    kWebAssemblyCompiledModule,
    kNumKinds,
} Kind;

extern void init();
extern IsolatePtr newIsolate();
extern void releaseIsolate(IsolatePtr ptr);

extern SandboxPtr newSandbox(IsolatePtr ptr);
extern void loadVM(SandboxPtr ptr, int vmType);
extern void releaseSandbox(SandboxPtr ptr);

extern ValueTuple Execute(SandboxPtr ptr, const char *code);
extern void setJSPath(SandboxPtr ptr, const char *jsPath);
extern void setSandboxGasLimit(SandboxPtr ptr, size_t gasLimit);

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
typedef char *(*globalGetFunc)(SandboxPtr, const char *, const char *, size_t *);
void InitGoStorage(putFunc, getFunc, delFunc,
    globalGetFunc);

//extern int goPut(SandboxPtr, char *, char *, size_t *);
//extern char *goGet(SandboxPtr, char *, size_t *);
//extern int goDel(SandboxPtr, char *, size_t *);

extern void goMapPut(SandboxPtr, const char *, const char *, const char *, size_t *);
extern void goMapGet(SandboxPtr, const char *, const char *, size_t *);
extern void goMapDel(SandboxPtr, const char *, const char *, size_t *);
extern void goMapKeys(SandboxPtr, const char *, size_t *);
extern void goMapLen(SandboxPtr, const char *, size_t *);
extern void goGlobalMapGet(SandboxPtr, const char *, const char *, const char *, size_t *);
extern void goGlobalMapKeys(SandboxPtr, const char *, const char *, size_t *);
extern void goGlobalMapLen(SandboxPtr, const char *, const char *, size_t *);

extern int compile(SandboxPtr, const char *code, const char **compiledCode);

#ifdef __cplusplus
}
#endif // __cplusplus

#endif // IOST_V8_ENGINE_H
