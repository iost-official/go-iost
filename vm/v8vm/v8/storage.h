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

    char* Put(const CStr key, const CStr value, const CStr owner);
    char* Has(const CStr key, const CStr owner, bool *result);
	char* Get(const CStr key, const CStr owner, CStr *result);
	char* Del(const CStr key, const CStr owner);
	char* MapPut(const CStr key, const CStr field, const CStr value, const CStr owner);
	char* MapHas(const CStr key, const CStr field, const CStr owner, bool *result);
	char* MapGet(const CStr key, const CStr field, const CStr owner, CStr *result);
	char* MapDel(const CStr key, const CStr field, const CStr owner);
	char* MapKeys(const CStr key, const CStr owner, CStr *result);
	char* MapLen(const CStr key, const CStr owner, size_t *result);
	
	char* GlobalHas(const CStr contract, const CStr key, const CStr owner, bool *result);
	char* GlobalGet(const CStr contract, const CStr key, const CStr owner, CStr *result);
	char* GlobalMapHas(const CStr contract, const CStr key, const CStr field, const CStr owner, bool *result);
	char* GlobalMapGet(const CStr contract, const CStr key, const CStr field, const CStr owner, CStr *result);
	char* GlobalMapKeys(const CStr contract,  const CStr key, const CStr owner, CStr *result);
	char* GlobalMapLen(const CStr contract, const CStr key, const CStr owner, size_t *result);

};

#endif // IOST_V8_STORAGE_H
