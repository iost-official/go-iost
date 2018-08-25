#!/bin/bash

find * -name "*.proto" | grep -v "vendor" | xargs -n1 \
protoc -I/usr/local/include -I. \
       -I$GOPATH/src/ \
       -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
       --gofast_out=plugins=grpc,paths=source_relative:.
