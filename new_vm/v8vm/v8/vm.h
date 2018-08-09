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
extern void releaseSandbox(SandboxPtr ptr);

extern ValueTuple Execute(SandboxPtr ptr, const char *code);

extern char *requireModule(SandboxPtr, char *);
extern int goPut(SandboxPtr, char *, char *, size_t *);
extern char *goGet(SandboxPtr, char *, size_t *);
extern int goDel(SandboxPtr, char *, size_t *);
extern void goMapPut(SandboxPtr, const char *, const char *, const char *, size_t *);
extern void goMapGet(SandboxPtr, const char *, const char *, size_t *);
extern void goMapDel(SandboxPtr, const char *, const char *, size_t *);
extern void goMapKeys(SandboxPtr, const char *, size_t *);
extern void goMapLen(SandboxPtr, const char *, size_t *);
extern void goGlobalGet(SandboxPtr, const char *, const char *, size_t *);
extern void goGlobalMapGet(SandboxPtr, const char *, const char *, const char *, size_t *);
extern void goGlobalMapKeys(SandboxPtr, const char *, const char *, size_t *);
extern void goGlobalMapLen(SandboxPtr, const char *, const char *, size_t *);
// blockchain
extern int goTransfer(SandboxPtr, char *, char *, char *, size_t *);
extern int goWithdraw(SandboxPtr, char *, char *, size_t *);
extern int goDeposit(SandboxPtr, char *, char *, size_t *);
extern int goTopUp(SandboxPtr, char *, char *, char *, size_t *);
extern int goCountermand(SandboxPtr, char *, char *, char *, size_t *);
extern char *goBlockInfo(SandboxPtr, size_t *);
extern char *goTxInfo(SandboxPtr, size_t *);
extern char *goCall(SandboxPtr, char *, char *, char *, size_t *);

#ifdef __cplusplus
}
#endif // __cplusplus

#endif // IOST_V8_ENGINE_H
