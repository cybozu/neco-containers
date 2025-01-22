SHELL:=/bin/bash

BIN_DIR := $(shell pwd)/bin

GO = go
GO_VET_OPTS = -v
GO_TEST_OPTS=-v -race

GO_FMT=gofmt
GO_FMT_OPTS=-s -l

CGO=CGO_ENABLED=0

DOCKER=docker
DOCKER_TAG=ghcr.io/cybozu/lbip-manager:dev

# Test tools
CUSTOM_CHECKER = $(BIN_DIR)/custom-checker
STATIC_CHECK = $(BIN_DIR)/staticcheck

$(CUSTOM_CHECKER):
	GOBIN=$(BIN_DIR) go install github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest

$(STATIC_CHECK):
	GOBIN=$(BIN_DIR) go install honnef.co/go/tools/cmd/staticcheck@latest


.PHONY: fmt
fmt:
	$(GO_FMT) $(GO_FMT_OPTS) .

.PHONY: mod
mod:
	$(GO) mod tidy

.PHONY: check-diff
check-diff: mod fmt
	git diff --exit-code --name-only

.PHONY: vet
vet:
	$(GO) vet $(GO_VET_OPTS) ./...

.PHONY: test
test: vet check-diff $(STATIC_CHECK) $(CUSTOM_CHECKER)
	$(STATIC_CHECK) ./...
	test -z "$$($(CUSTOM_CHECKER) -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	$(GO) test $(GO_TEST_OPTS) ./...

.PHONY: build
build: $(CMDS)
	$(CGO) $(GO) build $(GO_BUILD_OPT) -o $/lbip-manager

.PHONY: docker
docker:
	$(DOCKER) build -t $(DOCKER_TAG) .

.PHONY: clean
clean:
	-$(GO) clean
	-rm $(RM_OPTS) $(BIN_DIR)

.PHONY: all
all: test build docker

.DEFAULT_GOAL=all