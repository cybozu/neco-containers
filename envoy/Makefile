include ../tool.mk

SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	golangci-lint run
	git diff --exit-code --name-only

.PHONY: test
test:
	$(SUDO_GO) test -race -v ./...
