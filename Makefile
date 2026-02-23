BINARY  := asctl
MODULE  := $(shell go list -m)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w \
	-X '$(MODULE)/cmd.version=$(VERSION)' \
	-X '$(MODULE)/cmd.commit=$(COMMIT)' \
	-X '$(MODULE)/cmd.date=$(DATE)'

.PHONY: build clean install lint test

build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

install:
	go build -ldflags "$(LDFLAGS)" -o $(shell go env GOPATH)/bin/$(BINARY) .

clean:
	rm -rf bin

lint:
	go vet ./...

test:
	go test ./...
