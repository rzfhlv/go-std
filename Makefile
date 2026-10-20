CMD ?= rest
ARGS ?=
GO ?= go

.PHONY: all build test cover vet lint fmt fmt-check tidy clean install-tools run help

all: build

## build: compile all commands (binaries go to bin/)
build:
	@mkdir -p bin
	$(GO) build -o bin/ ./...

## run: run a command, e.g. `make run CMD=rest` or `make run CMD=cli ARGS="list"`
run:
	$(GO) run ./cmd/$(CMD) $(ARGS)

## test: run all tests with race detection
test:
	$(GO) test -race ./...

## cover: run tests with coverage report
cover:
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

## vet: run go vet
vet:
	$(GO) vet ./...

## lint: run golangci-lint
lint:
	golangci-lint config verify
	golangci-lint run ./...

## fmt: format code with gofumpt (applies changes)
fmt:
	gofumpt -w .

## fmt-check: verify formatting without changing files (CI-safe)
fmt-check:
	@fmt_out=$$(gofumpt -l .); \
	if [ -n "$$fmt_out" ]; then \
		echo "Files need formatting:"; echo "$$fmt_out"; exit 1; \
	fi

## tidy: sync go.mod / go.sum
tidy:
	$(GO) mod tidy

## install-tools: install lint/format tooling
install-tools:
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

## clean: remove build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //'
