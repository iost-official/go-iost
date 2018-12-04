#ifndef IOST_V8_COMPILE_H
#define IOST_V8_COMPILE_H

#include "sandbox.h"

int compile(SandboxPtr, const char *code, const char **compiledCode);
int validate(SandboxPtr ptr, const char *code, const char *abi, const char **result);
CustomStartupData createStartupData();
CustomStartupData createCompileStartupData();

extern intptr_t externalRef[];

#endif // IOST_V8_COMPILE_H