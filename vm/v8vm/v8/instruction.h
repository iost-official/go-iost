#ifndef IOST_V8_INSTRUCTION_H
#define IOST_V8_INSTRUCTION_H

#include "sandbox.h"
#include "stddef.h"

void InitInstruction(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTContractInstruction(const FunctionCallbackInfo<Value> &info);

class IOSTContractInstruction {
private:
    SandboxPtr sbxPtr;
public:
    IOSTContractInstruction(SandboxPtr ptr): sbxPtr(ptr) {}

    size_t Incr(size_t num) {
        Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
        sbx->gasUsed += num;
        return sbx->gasUsed;
    }
    size_t Count() {
        Sandbox *sbx = static_cast<Sandbox*>(sbxPtr);
        return sbx->gasUsed;
    }
};

#endif // IOST_V8_INSTRUCTION_H