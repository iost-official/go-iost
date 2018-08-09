#ifndef IOST_V8_SANDBOX_H
#define IOST_V8_SANDBOX_H

#include "v8.h"
#include "vm.h"

using namespace v8;

extern ValueTuple Execution(SandboxPtr ptr, const char *code);

#endif // IOST_V8_SANDBOX_H