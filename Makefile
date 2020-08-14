VERSION=$(shell git describe --tags --abbrev=0)-$(shell git rev-parse --short HEAD)
GOFLAGS=-ldflags "-w -s -X $(shell go list)/cmd.version=$(VERSION)"
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GOFLAGS)

install:
	go install $(GOFLAGS)

dist:
	goreleaser releaser --skip-publish --snapshot

service:
	edu service --install

snapshot:
	goreleaser release --skip-publish --snapshot

clean:
	go clean
	$(RM) -r dist

.PHONY: build install clean
