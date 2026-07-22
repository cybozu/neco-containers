GO=go
GO_TEST_OPTS=-v -race

DOCKER=docker
DOCKER_FILE=./Dockerfile
DOCKER_REPO=ghcr.io/cybozu
DOCKER_TAG=dev
DOCKER_CONTEXT=.

include ../tool.mk

.PHONY: mod
mod:
	$(GO) mod tidy

.PHONY: check-diff
check-diff: mod
	aqua install
	golangci-lint run
	git diff --exit-code --name-only

.PHONY: test
test:
	$(GO) test $(GO_TEST_OPTS) ./...

.PHONY: build
build:
	$(GO) build $(GO_BUILD_OPT) -o ./bin/tcp-keepalive .

.PHONY: docker
docker:
	$(DOCKER) image build -f $(DOCKER_FILE) -t $(DOCKER_REPO)/tcp-keepalive:$(DOCKER_TAG) $(DOCKER_CONTEXT)

.PHONY: clean
clean:
	-$(GO) clean
	-rm $(RM_OPTS) ./bin/*

.PHONY: all
all: check-diff test build

.DEFAULT_GOAL=all
