#include "sandbox.h"

SnapshotData *createSnapshotData(const char *script) {
    StartupData data = V8::CreateSnapshotDataBlob(script);
    return new SnapshotData{data.data, data.size};
}
