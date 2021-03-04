#!/bin/bash

if [ $(uname) == Darwin ]
then
    export CGO_LDFLAGS=-L$(pwd)/vm/v8vm/v8/libv8/_darwin_amd64
    export CGO_CFLAGS=-I$(pwd)/vm/v8vm/v8/include/_darwin_amd64
    export DYLD_LIBRARY_PATH=$(pwd)/vm/v8vm/v8/libv8/_darwin_amd64
fi

if [ $(uname) == Linux ]
then
    export CGO_LDFLAGS=-L$(pwd)/vm/v8vm/v8/libv8/_linux_amd64
    export CGO_CFLAGS=-I$(pwd)/vm/v8vm/v8/include/_linux_amd64
    export LD_LIBRARY_PATH=$(pwd)/vm/v8vm/v8/libv8/_linux_amd64
fi

# go install github.com/golang/protobuf/protoc-gen-go@v1.4.3
# go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0 github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0

find * -name "*.proto" | grep -v "vendor" | grep -v "third_party" | xargs -n1 \
protoc -I/usr/local/include -I. -Ivendor \
       -Iproto/third_party/googleapis \
       --go_out=plugins=grpc,paths=source_relative:.

protoc -I/usr/local/include -I. -Ivendor \
       -Iproto/third_party/googleapis \
       --grpc-gateway_out=logtostderr=true,paths=source_relative:.\
       rpc/pb/rpc.proto
