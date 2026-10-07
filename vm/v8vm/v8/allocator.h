//
// Created by Shiqi Liu on 2018/10/29.
//

#ifndef GO_IOST_ALLOCATOR_H
#define GO_IOST_ALLOCATOR_H

#include <stdint.h>
#include <v8.h>

using namespace v8;
class ArrayBufferAllocator : public ArrayBuffer::Allocator {
public:
    ArrayBufferAllocator();
    virtual ~ArrayBufferAllocator();

    /**
     * Allocate |length| bytes. Return NULL if allocation is not successful.
     * Memory should be initialized to zeroes.
     */
    virtual void *Allocate(size_t length);

    /**
     * Allocate |length| bytes. Return NULL if allocation is not successful.
     * Memory does not have to be initialized.
     */
    virtual void *AllocateUninitialized(size_t length);

    /**
     * Free the memory block of size |length|, pointed to by |data|.
     * That memory is guaranteed to be previously allocated by |Allocate|.
     */
    virtual void Free(void *data, size_t length);

    size_t GetCurrentAllocatedMemSize();

    size_t GetMaxAllocatedMemSize();

private:
    size_t current_allocated_size;
    size_t max_allocated_size;

    void AddAllocatedSize(size_t length);
};

#endif //GO_IOST_ALLOCATOR_H
