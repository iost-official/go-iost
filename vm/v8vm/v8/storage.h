#ifndef IOST_V8_STORAGE_H
#define IOST_V8_STORAGE_H

#include "sandbox.h"
#include "stddef.h"

using namespace v8;

void InitStorage(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTContractStorage(const FunctionCallbackInfo<Value> &info);

class IOSTContractStorage {
private:
    SandboxPtr sbxPtr;
public:
    IOSTContractStorage(SandboxPtr ptr): sbxPtr(ptr) {}

    char* Put(const char *key, const char *value, const char *owner);
    char* Has(const char *key, const char *owner, bool *result);
	char* Get(const char *key, const char *owner, char **result);
	char* Del(const char *key, const char *owner);
	char* MapPut(const char *key, const char *field, const char *value, const char *owner);
	char* MapHas(const char *key, const char *field, const char *owner, bool *result);
	char* MapGet(const char *key, const char *field, const char *owner, char **result);
	char* MapDel(const char *key, const char *field, const char *owner);
	char* MapKeys(const char *key, const char *owner, char **result);
	char* MapLen(const char *key, const char *owner, size_t *result);
	
	char* GlobalHas(const char *contract, const char *key, const char *owner, bool *result);
	char* GlobalGet(const char *contract, const char *key, const char *owner, char **result);
	char* GlobalMapHas(const char *contract, const char *key, const char *field, const char *owner, bool *result);
	char* GlobalMapGet(const char *contract, const char *key, const char *field, const char *owner, char **result);
	char* GlobalMapKeys(const char *contract,  const char *key, const char *owner, char **result);
	char* GlobalMapLen(const char *contract, const char *key, const char *owner, size_t *result);

};

#endif // IOST_V8_STORAGE_H
