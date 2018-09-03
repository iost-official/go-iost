#ifndef IOST_V8_CONSOLE_H
#define IOST_V8_CONSOLE_H

#include "sandbox.h"

void InitConsole(Isolate *isolate, Local<ObjectTemplate> globalTpl);

#endif // IOST_V8_CONSOLE_H