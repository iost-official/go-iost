#ifndef IOST_V8_SANDBOX_H
#define IOST_V8_SANDBOX_H

#include "v8.h"
#include "vm.h"
#include "ThreadPool.h"
#include "allocator.h"
#include <string>

using namespace v8;

#ifndef NewCStr
#define NewCStr(name, str) \
  v8::String::Utf8Value __v8String_ ## name(str);\
  CStr name = {*__v8String_ ## name, __v8String_ ## name.length()}
#endif

#ifndef NewCStrChecked
#define INPUT_MAX_LENGTH  65536
#define NewCStrChecked(name, str, isolate) \
  v8::Local<String> __v8LocalString_ ## name = v8::Local<String>::Cast(str);\
  if (__v8LocalString_ ## name->Length() > INPUT_MAX_LENGTH) {\
    Local<Value> __v8LocalStringError_ ## name = Exception::Error(\
        String::NewFromUtf8(isolate, "input string too long")\
    );\
    isolate->ThrowException(__v8LocalStringError_ ## name);\
    return;\
  }\
  NewCStr(name, str)
#endif

typedef struct {
  Persistent<Context> context;
  Isolate *isolate;
  ArrayBufferAllocator* allocator;
  const char *jsPath;
  size_t gasUsed;
  size_t gasLimit;
  size_t memLimit;
  std::unique_ptr<ThreadPool> threadPool;
  void* blockchain;
  void* crypto;
  void* instruction;
  void* storage;
} Sandbox;

extern ValueTuple Execution(SandboxPtr ptr, const CStr code, long long int expireTime);

size_t MemoryUsage(Isolate* isolate, ArrayBufferAllocator* allocator);

std::string reportException(Isolate *isolate, Local<Context> ctx, TryCatch& tryCatch);

#endif // IOST_V8_SANDBOX_H