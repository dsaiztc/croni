SHELL := /bin/bash
.DEFAULT_GOAL := build

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo "")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/dsaiztc/croni/cmd.version=$(VERSION) \
           -X github.com/dsaiztc/croni/cmd.commit=$(COMMIT) \
           -X github.com/dsaiztc/croni/cmd.date=$(DATE)

.PHONY: build test vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/croni .

test:
	go test -race ./...

vet:
	go vet ./...

clean:
	rm -rf bin/ dist/
