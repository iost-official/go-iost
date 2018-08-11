#ifndef IOST_V8_STORAGE_H
#define IOST_V8_STORAGE_H

#include "v8.h"
#include "vm.h"
#include "stddef.h"

using namespace v8;

void InitStorage(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTContractStorage(const FunctionCallbackInfo<Value> &info);

class IOSTContractStorage {
private:
    SandboxPtr sbx;
public:
    IOSTContractStorage(SandboxPtr ptr): sbx(ptr) {}

    int Put(char *key, char *value);
    char *Get(char *key);
    int Del(char *key);
    void MapPut(const char *key, const char *field, const char *value) {
//        size_t gasUsed = 0;
//        char *ret = goMapPut(sbx, key, field, value, &gasUsed);
//        return ret;
    }
    void MapGet(const char *key, const char *field) {
//        size_t gasUsed = 0;
//        char *ret = goMapGet(sbx, key, field, &gasUsed);
//        return ret;
    }
    void MapDel(const char *key, const char *field) {
//        size_t gasUsed = 0;
//        char *ret = goMapDel(sbx, key, field, &gasUsed);
//        return ret;
    }
    void MapKeys(const char *key) {
//        size_t gasUsed = 0;
//        char *ret = goMapKeys(sbx, key, &gasUsed);
//        return ret;
    }
    void MapLen(const char *key) {
//        size_t gasUsed = 0;
//        char *ret = goMapLen(sbx, key, &gasUsed);
//        return ret;
    }
    void GlobalGet(const char *contract, const char *key) {
//        size_t gasUsed = 0;
//        char *ret = goGlobalGet(sbx, contract, key, &gasUsed);
//        return ret;
    }
    void GlobalMapGet(const char *contract, const char *key, const char *field) {
//        size_t gasUsed = 0;
//        char *ret = goGlobalMapGet(sbx, contract, key, field, &gasUsed);
//        return ret;
    }
    void GlobalMapKeys(const char *contract, const char *key) {
//        size_t gasUsed = 0;
//        char *ret = goGlobalMapKeys(sbx, contract, key, &gasUsed);
//        return ret;
    }
    void GlobalMapLen(const char *contract, const char *key) {
//        size_t gasUsed = 0;
//        char *ret = goGlobalMapLen(sbx, contract, key, &gasUsed);
//        return ret;
    }
};

#endif // IOST_V8_STORAGE_H