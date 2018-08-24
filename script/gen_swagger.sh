#!/bin/bash

protoc -I/usr/local/include -I. \
       -I$GOPATH/src \
       -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
       --swagger_out=logtostderr=true:. \
       rpc/apis.proto
