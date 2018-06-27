GO = go

VERSION = 1.0.0
COMMIT = $(shell git rev-parse --short HEAD)
PROJECT = github.com/iost-official/prototype
DOCKER_IMAGE = iost-node:$(VERSION)-$(COMMIT)
TARGET_DIR = build

.PHONY: all build iserver iwallet register txsender lint image devimage install clean

all: build

build: iserver iwallet register txsender

iserver:
	$(GO) build -o $(TARGET_DIR)/iserver $(PROJECT)/iserver

iwallet:
	$(GO) build -o $(TARGET_DIR)/iwallet $(PROJECT)/iwallet

register:
	$(GO) build -o $(TARGET_DIR)/register $(PROJECT)/network/main/

txsender:
	$(GO) build -o $(TARGET_DIR)/txsender $(PROJECT)/txsender

lint:
	@gometalinter --config=.gometalinter.json ./...

image: devimage
	docker run --rm -v `pwd`:/gopath/src/github.com/iost-official/prototype iost-dev make
	docker build -f Dockerfile.run -t $(DOCKER_IMAGE) .

devimage:
	docker build -f Dockerfile.dev -t iost-dev .

install:
	go install ./iwallet/
	go install ./iserver/

clean:
	rm -f ${TARGET_DIR}
