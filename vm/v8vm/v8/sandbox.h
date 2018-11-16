#ifndef IOST_V8_SANDBOX_H
#define IOST_V8_SANDBOX_H

#include "v8.h"
#include "vm.h"
#include "ThreadPool.h"

using namespace v8;

typedef struct {
  Persistent<Context> context;
  Isolate *isolate;
  const char *jsPath;
  size_t gasUsed;
  size_t gasLimit;
  std::unique_ptr<ThreadPool> threadPool;
} Sandbox;

extern ValueTuple Execution(SandboxPtr ptr, const char *code, long long int expireTime);

#endif // IOST_V8_SANDBOX_H