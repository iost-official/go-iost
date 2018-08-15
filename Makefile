GO = go

VERSION = 1.0.3
COMMIT = $(shell git rev-parse --short HEAD)
PROJECT = github.com/iost-official/Go-IOS-Protocol
DOCKER_IMAGE = iost-node:$(VERSION)-$(COMMIT)
TARGET_DIR = target

ifeq ($(shell uname),Darwin)
	export CGO_LDFLAGS=-L$(shell pwd)/new_vm/v8vm/v8/libv8/_darwin_amd64
	export CGO_CFLAGS=-I$(shell pwd)/new_vm/v8vm/v8/include/_darwin_amd64
	export DYLD_LIBRARY_PATH=$(shell pwd)/new_vm/v8vm/v8/libv8/_darwin_amd64
endif

ifeq ($(shell uname),Linux)
	export CGO_LDFLAGS=-L$(shell pwd)/new_vm/v8vm/v8/libv8/_linux_amd64
	export CGO_CFLAGS=-I$(shell pwd)/new_vm/v8vm/v8/include/_linux_amd64
	export LD_LIBRARY_PATH=$(shell pwd)/new_vm/v8vm/v8/libv8/_linux_amd64
endif

.PHONY: all build iserver iwallet lint test image devimage install clean

all: build

build: iserver iwallet

iserver:
	$(GO) build -o $(TARGET_DIR)/iserver $(PROJECT)/cmd/iserver

iwallet:
	$(GO) build -o $(TARGET_DIR)/iwallet $(PROJECT)/cmd/iwallet

lint:
	@gometalinter --config=.gometalinter.json ./...

test:
	@go test $(shell go list ./... | grep -vE "/new_vm/v8vm|/core/new_blockcache")

image: devimage
	docker run --rm -v `pwd`:/gopath/src/github.com/iost-official/Go-IOS-Protocol iost-dev make
	docker build -f Dockerfile.run -t $(DOCKER_IMAGE) .

devimage:
	docker build -f Dockerfile.dev -t iost-dev .

install:
	go install ./cmd/iwallet/
	go install ./cmd/iserver/

clean:
	rm -f ${TARGET_DIR}
