VERSION=$(shell git describe --tags --abbrev=0)-$(shell git rev-parse --short HEAD)
GOFLAGS=-ldflags "-X $(shell go list)/cmd.version=$(VERSION)"

build:
	go build $(GOFLAGS)

install:
	go install $(GOFLAGS)

dist:
	goreleaser releaser --skip-publish --snapshot

service: misc/systemd/edu.service
	@if [ ! -d misc/systemd ]; then mkdir -p misc/systemd; fi
	@if systemctl status edu > /dev/null 1>&2; then systemctl stop edu; fi
	install $< /etc/systemd/system
	systemctl enable edu

misc/systemd/edu.service:
	go build
	./edu gen-service -f $@
	go clean

clean:
	go clean
	$(RM) -r dist

.PHONY: build install clean
