SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: lint check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint:
	aqua install
	golangci-lint run

.PHONY: test
test:
	$(SUDO_GO) test -race -v ./...

include ../tool.mk
