#ifndef IOST_V8_CONSOLE_H
#define IOST_V8_CONSOLE_H

#include "sandbox.h"

// This Class Provide Console.Log Function so JS code can use Go log.
void InitConsole(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewConsoleLog(const FunctionCallbackInfo<Value> &args);

#endif // IOST_V8_CONSOLE_H