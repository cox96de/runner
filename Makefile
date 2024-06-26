export SHELL := /bin/bash

.PHONY: format
format:
	gofmt -w .
	go mod tidy
#gofumpt is more strict than gofmt
	go run mvdan.cc/gofumpt@latest -l -w .

.PHONY: lint
lint:
	golangci-lint run --new-from-rev=origin/master --timeout=10m --go=1.21
build: build_executor build_server build_agent
build_executor:
	mkdir -p output
	CGO_ENABLED=0 go build -o output/executor ./cmd/executor/
build_server:
	mkdir -p output
	CGO_ENABLED=0 go build -o output/server ./cmd/server/

build_agent:
	mkdir -p output
	CGO_ENABLED=0 go build -o output/agent ./cmd/agent/
