#include "vm.h"
#include "v8.h"
#include "allocator.h"
#include "sandbox.h"
#include "compile.h"
#include "snapshot_blob.bin.h"
#include "natives_blob.bin.h"
#include "iostream"

#include "libplatform/libplatform.h"

#include <assert.h>
#include <stdlib.h>
#include <stdio.h>

using namespace v8;

void init() {
    /*
    std::string noGC ("--expose_gc");
    std::cout << "Start V8 With " << noGC << std::endl;
    V8::SetFlagsFromString(noGC.c_str(), noGC.length()+1);
    */
    V8::InitializeICU();

    Platform *platform = platform::CreateDefaultPlatform();
    V8::InitializePlatform(platform);
#ifdef __linux__
    StartupData nativesData, snapshotData;
    nativesData.data = reinterpret_cast<char *>(natives_blob_bin);
    nativesData.raw_size = natives_blob_bin_len;
    snapshotData.data = reinterpret_cast<char *>(snapshot_blob_bin);
    snapshotData.raw_size = snapshot_blob_bin_len;
    V8::SetNativesDataBlob(&nativesData);
    V8::SetSnapshotDataBlob(&snapshotData);
#endif
    V8::Initialize();
    return;
}

IsolateWrapperPtr newIsolate(CustomStartupData customStartupData) {
    Isolate::CreateParams params;
    IsolateWrapper* isolateWrapperPtr = new IsolateWrapper();

    StartupData* blob = new StartupData;
    blob->data = customStartupData.data;
    blob->raw_size = customStartupData.raw_size;

    extern intptr_t externalRef[];
    params.snapshot_blob = blob;
    params.array_buffer_allocator = new ArrayBufferAllocator();
    params.external_references = externalRef;
    isolateWrapperPtr->isolate = static_cast<Isolate*>(Isolate::New(params));
    isolateWrapperPtr->allocator = static_cast<void*>(params.array_buffer_allocator);
    return isolateWrapperPtr;
}

void releaseIsolate(IsolateWrapperPtr ptr) {
    if (ptr == nullptr) {
        return;
    }

    Isolate *isolate = static_cast<Isolate*>((static_cast<IsolateWrapper*>(ptr))->isolate);
    isolate->Dispose();
    return;
}

void lowMemoryNotification(IsolateWrapperPtr ptr) {
    if (ptr == nullptr) {
        return;
    }

    Isolate *isolate = static_cast<Isolate*>((static_cast<IsolateWrapper*>(ptr))->isolate);
    isolate->ContextDisposedNotification();
    isolate->LowMemoryNotification();
    //isolate->RequestGarbageCollectionForTesting(Isolate::GarbageCollectionType::kFullGarbageCollection);

    return;
}

ValueTuple Execute(SandboxPtr ptr, const char *code, long long int expireTime) {
    ValueTuple ret = Execution(ptr, code, expireTime);
    return ret;
}
