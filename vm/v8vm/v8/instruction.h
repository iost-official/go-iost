#ifndef IOST_V8_INSTRUCTION_H
#define IOST_V8_INSTRUCTION_H

#include "sandbox.h"
#include "stddef.h"
#include <iostream>
#include <cstring>
#include <string>

void InitInstruction(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTContractInstruction(const FunctionCallbackInfo<Value> &info);
void IOSTContractInstruction_Count(const FunctionCallbackInfo<Value> &args);
void IOSTContractInstruction_Incr(const FunctionCallbackInfo<Value> &args);

class IOSTContractInstruction {
private:
    Sandbox* sbxPtr;
    Isolate* isolate;
    int count;
public:
    IOSTContractInstruction(SandboxPtr ptr){
        sbxPtr = static_cast<Sandbox*>(ptr);
        isolate = sbxPtr->isolate;
        count = 0;
    }

    size_t Incr(size_t num) {
        if (sbxPtr->gasUsed > SIZE_MAX - num) {
            Local<Value> err = Exception::Error(
                String::NewFromUtf8(isolate, "IOSTContractInstruction_Incr gas overflow size_t")
            );
            isolate->ThrowException(err);
            return 0;
        }

        sbxPtr->gasUsed += num;
        count ++;
        return sbxPtr->gasUsed;
    }
    size_t Count() {
        return count;
    }
    void MemUsageCheck(){
        size_t usedMem = MemoryUsage(isolate, sbxPtr->allocator);
        if (usedMem > sbxPtr->memLimit){
            Local<Value> err = Exception::Error(
                String::NewFromUtf8(isolate, ("IOSTContractInstruction_Incr Memory Using too much! used: " + std::to_string(usedMem) + " Limit: " + std::to_string(sbxPtr->memLimit)).c_str())
            );
            isolate->ThrowException(err);
        }
        return;
    }
};

#endif // IOST_V8_INSTRUCTION_H