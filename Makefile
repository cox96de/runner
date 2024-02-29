export SHELL := /bin/bash

.PHONY: format
format:
	gofmt -w .
	go mod tidy
#gofumpt is more strict than gofmt
	go run mvdan.cc/gofumpt@latest -l -w .

.PHONY: lint
lint:
# macvm excluded, there's some problem.
	golangci-lint run --new-from-rev=origin/master --timeout=10m --go=1.22 --skip-dirs  "(^|/)macvm($|/)"
build: build_executor build_server
build_executor:
	mkdir -p output
	CGO_ENABLED=0 go build -o output/executor ./cmd/executor...
build_server:
	mkdir -p output
	CGO_ENABLED=0 go build -o output/server ./cmd/server...
