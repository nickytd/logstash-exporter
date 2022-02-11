GO              ?= go
GOPATH          := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOLINTER        ?= $(GOPATH)/bin/golangci-lint
GOLINTERCONFIG  ?= .golangci.yml
pkgs            = $(shell $(GO) list ./... | grep -v /vendor/)
TARGET          ?= logstash_exporter

PREFIX          ?= $(shell pwd)
BIN_DIR         ?= $(shell pwd)

all: clean format lint build test

test:
	@echo ">> running tests"
	@$(GO) test -short $(pkgs)

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

lint: $(GOPATH)/bin/golangci-lint
	@echo ">> linting code"
	@$(GOLINTER) run -c $(GOLINTERCONFIG) --sort-results

build: format lint
	@echo ">> building binaries"
	@$(GO) mod download
	@$(GO) build $(PREFIX)

clean:
	@echo ">> Cleaning up"
	@find . -type f -name '*~' -exec rm -fv {} \;
	@rm -fv $(TARGET)

$(GOPATH)/bin/golangci-lint:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.0

.PHONY: all format build test clean $(GOPATH)/bin/golangci-lint lint
