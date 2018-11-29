#ifndef IOST_V8_SANDBOX_H
#define IOST_V8_SANDBOX_H

#include "v8.h"
#include "vm.h"
#include "ThreadPool.h"
#include "allocator.h"

using namespace v8;

typedef struct {
  Persistent<Context> context;
  Isolate *isolate;
  ArrayBufferAllocator* allocator;
  const char *jsPath;
  size_t gasUsed;
  size_t gasLimit;
  size_t memLimit;
  std::unique_ptr<ThreadPool> threadPool;
} Sandbox;

extern ValueTuple Execution(SandboxPtr ptr, const char *code, long long int expireTime);

size_t MemoryUsage(Isolate* isolate, ArrayBufferAllocator* allocator);

#endif // IOST_V8_SANDBOX_H