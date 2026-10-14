#!/bin/bash

set -eo pipefail

if [ $(uname) == Darwin ]; then
	export CGO_LDFLAGS=-L$(pwd)/vm/v8vm/v8/libv8/_darwin_amd64
	export CGO_CFLAGS=-I$(pwd)/vm/v8vm/v8/include/_darwin_amd64
	export DYLD_LIBRARY_PATH=$(pwd)/vm/v8vm/v8/libv8/_darwin_amd64
fi

if [ $(uname) == Linux ]; then
	export CGO_LDFLAGS=-L$(pwd)/vm/v8vm/v8/libv8/_linux_amd64
	export CGO_CFLAGS=-I$(pwd)/vm/v8vm/v8/include/_linux_amd64
	export LD_LIBRARY_PATH=$(pwd)/vm/v8vm/v8/libv8/_linux_amd64
fi


function install_tools() {
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.31.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.31.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
}

function generate_proto_and_grpc() {
	# Exit 255 makes xargs stop on the first failed generation.
	find * -name "*.proto" | grep -v "third_party" | xargs -n1 sh -c 'protoc "$@" || exit 255' sh \
		-I/usr/local/include -I. \
		-Iproto/third_party/googleapis \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false,paths=source_relative
}

function generate_grpc_gateway() {
	protoc -I/usr/local/include -I. \
		-Iproto/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true,paths=source_relative:. rpc/pb/rpc.proto
}

function generate_openapi() {
	protoc -I/usr/local/include -I. \
		-Iproto/third_party/googleapis \
		--openapiv2_out=logtostderr=true:. \
		rpc/pb/rpc.proto
}

install_tools
generate_proto_and_grpc
generate_grpc_gateway
generate_openapi
