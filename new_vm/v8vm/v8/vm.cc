#include "vm.h"
#include "v8.h"
#include "sandbox.h"
#include "snapshot_blob.bin.h"
#include "natives_blob.bin.h"

#include "libplatform/libplatform.h"

#include <assert.h>
#include <stdlib.h>
#include <stdio.h>

using namespace v8;

void init() {
    V8::InitializeICU();

    Platform *platform = platform::CreateDefaultPlatform();
    V8::InitializePlatform(platform);

    StartupData nativesData, snapshotData;
    nativesData.data = reinterpret_cast<char *>(natives_blob_bin);
    nativesData.raw_size = natives_blob_bin_len;
    snapshotData.data = reinterpret_cast<char *>(snapshot_blob_bin);
    snapshotData.raw_size = snapshot_blob_bin_len;
    V8::SetNativesDataBlob(&nativesData);
    V8::SetSnapshotDataBlob(&snapshotData);

    V8::Initialize();
    return;
}

IsolatePtr newIsolate() {
  Isolate::CreateParams create_params;
  create_params.array_buffer_allocator = ArrayBuffer::Allocator::NewDefaultAllocator();
  return static_cast<IsolatePtr>(Isolate::New(create_params));
}

void releaseIsolate(IsolatePtr ptr) {
    if (ptr == nullptr) {
        return;
    }

    Isolate *isolate = static_cast<Isolate*>(ptr);
    isolate->Dispose();
    return;
}

ValueTuple Execute(SandboxPtr ptr, const char *code) {
    ValueTuple ret = Execution(ptr, code);
    return ret;
}
