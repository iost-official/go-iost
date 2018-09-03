#ifndef IOST_V8_CONSOLE_H
#define IOST_V8_CONSOLE_H

#include "sandbox.h"

// This Class Provide Console.Log Function so JS code can use Go log.
void InitConsole(Isolate *isolate, Local<ObjectTemplate> globalTpl);

#endif // IOST_V8_CONSOLE_H