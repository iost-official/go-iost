GO = go

VERSION = 1.0.3
COMMIT = $(shell git rev-parse --short HEAD)
PROJECT = github.com/iost-official/Go-IOS-Protocol
DOCKER_IMAGE = iost-node:$(VERSION)-$(COMMIT)
TARGET_DIR = target

.PHONY: all build iserver iwallet lint image devimage install clean

all: build

build: iserver iwallet

iserver:
	$(GO) build -o $(TARGET_DIR)/iserver $(PROJECT)/cmd/iserver

iwallet:
	$(GO) build -o $(TARGET_DIR)/iwallet $(PROJECT)/cmd/iwallet

lint:
	@gometalinter --config=.gometalinter.json ./...

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
