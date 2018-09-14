GO = go

VERSION = 2.0.0
COMMIT = $(shell git rev-parse --short HEAD)
PROJECT = github.com/iost-official/Go-IOS-Protocol
DOCKER_IMAGE = iost-node:$(VERSION)-$(COMMIT)
TARGET_DIR = target

ifeq ($(shell uname),Darwin)
	export CGO_LDFLAGS=-L$(shell pwd)/vm/v8vm/v8/libv8/_darwin_amd64 -lvm
	export CGO_CFLAGS=-I$(shell pwd)/vm/v8vm/v8/include/_darwin_amd64
	export DYLD_LIBRARY_PATH=$(shell pwd)/vm/v8vm/v8/libv8/_darwin_amd64
endif

ifeq ($(shell uname),Linux)
	export CGO_LDFLAGS=-L$(shell pwd)/vm/v8vm/v8/libv8/_linux_amd64 -lvm -lv8
	export CGO_CFLAGS=-I$(shell pwd)/vm/v8vm/v8/include/_linux_amd64
	export LD_LIBRARY_PATH=$(shell pwd)/vm/v8vm/v8/libv8/_linux_amd64
endif

.PHONY: all build iserver iwallet lint test image devimage swagger protobuf install clean debug clear_debug_file

all: build

build: iserver iwallet

iserver:
	$(GO) build -o $(TARGET_DIR)/iserver $(PROJECT)/cmd/iserver

iwallet:
	$(GO) build -o $(TARGET_DIR)/iwallet $(PROJECT)/cmd/iwallet

lint:
	@gometalinter --config=.gometalinter.json ./...

test:
	

image:
	docker run --rm -v `pwd`:/gopath/src/github.com/iost-official/Go-IOS-Protocol iostio/iost-dev make
	docker build -f Dockerfile.run -t $(DOCKER_IMAGE) .

devimage:
	docker build -f Dockerfile.dev -t iostio/iost-dev .

swagger:
	./script/gen_swagger.sh

protobuf:
	./script/gen_protobuf.sh

install:
	go install ./cmd/iwallet/
	go install ./cmd/iserver/

clean:
	rm -f ${TARGET_DIR}

debug: build
	target/iserver -f config/iserver.yml

clear_debug_file:
	rm -rf StatePoolDB/
	rm -rf leveldb/
	rm priv.key
	rm routing.table
