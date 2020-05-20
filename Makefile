VERSION=$(shell git describe --tags --abbrev=0)-$(shell git rev-parse --short HEAD)
GOFLAGS=-ldflags "-X $(shell go list)/cmd.version=$(VERSION)"

build:
	go build $(GOFLAGS)

install:
	go install $(GOFLAGS)

clean:
	go clean
