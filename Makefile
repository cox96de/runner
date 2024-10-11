export SHELL := /bin/bash
DOCKER_IMG_BASE:=cox96de
.PHONY: format
format:
	gofmt -w .
	go mod tidy
	docker run --volume "$$(pwd):/workspace" --workdir /workspace bufbuild/buf format -w
#gofumpt is more strict than gofmt
	go run mvdan.cc/gofumpt@latest -l -w .

.PHONY: lint
lint:
	golangci-lint run --new-from-rev=origin/master --timeout=10m --go=1.23
build: build_executor build_server build_agent
build_executor:
	mkdir -p output
	CGO_ENABLED=0 go build -o output/executor ./cmd/executor/
build_server:
	mkdir -p output
	go build -o output/server ./cmd/server/

build_agent:
	mkdir -p output
	CGO_ENABLED=0 go build -o output/agent ./cmd/agent/
build_docker: build_agent_docker build_agent_debian_docker build_executor_docker build_server_docker build_simplecli_docker

build_server_docker:
	docker build -t $(DOCKER_IMG_BASE)/runner-server -f build/server.Dockerfile .
build_executor_docker:
	docker build -t $(DOCKER_IMG_BASE)/runner-executor -f build/executor.Dockerfile .
build_agent_docker:
	docker build -t $(DOCKER_IMG_BASE)/runner-agent -f build/agent.Dockerfile .
build_agent_debian_docker:
	docker build -t $(DOCKER_IMG_BASE)/runner-agent-debian -f build/agent.debian.Dockerfile .
build_vm_runtime:
	docker build -t $(DOCKER_IMG_BASE)/runner-vm-runtime -f engine/vm/runtime/Dockerfile .
build_simplecli_docker:
	docker build -t $(DOCKER_IMG_BASE)/runner-simplecli -f build/simplecli.Dockerfile .