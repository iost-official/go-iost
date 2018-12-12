#ifndef IOST_V8_COMPILE_H
#define IOST_V8_COMPILE_H

#include "sandbox.h"

int compile(SandboxPtr, const CStr code, CStr *compiledCode, CStr *errMsg);
int validate(SandboxPtr ptr, const CStr code, const CStr abi, CStr *result, CStr *errMsg);
CustomStartupData createStartupData();
CustomStartupData createCompileStartupData();

extern intptr_t externalRef[];

#endif // IOST_V8_COMPILE_H