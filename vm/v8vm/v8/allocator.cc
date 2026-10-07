//
// Created by Shiqi Liu on 2018/10/29.
//
#include "allocator.h"

ArrayBufferAllocator::ArrayBufferAllocator(){
    this->current_allocated_size = 0;
    this->max_allocated_size = 0;
}

ArrayBufferAllocator::~ArrayBufferAllocator(){}
void* ArrayBufferAllocator::Allocate(size_t length) {
    this->AddAllocatedSize(length);
    void* data = calloc(length, 1);
    return data;
}

void* ArrayBufferAllocator::AllocateUninitialized(size_t length) {
    this->AddAllocatedSize(length);
    void* data = malloc(length);
    return data;
}

void ArrayBufferAllocator::Free(void* data, size_t length) {
    this->current_allocated_size -= length;
    free(data);
}

size_t ArrayBufferAllocator::GetCurrentAllocatedMemSize(){
    return this->current_allocated_size;
}

size_t ArrayBufferAllocator::GetMaxAllocatedMemSize(){
    return this->max_allocated_size;
}

void ArrayBufferAllocator::AddAllocatedSize(size_t length){
    this->current_allocated_size += length;
    if (this->current_allocated_size > this->max_allocated_size)
        this->max_allocated_size = this->current_allocated_size;
}
