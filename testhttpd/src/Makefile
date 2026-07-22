SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: test
test:
	aqua install
	golangci-lint run
	$(SUDO_GO) test -race -v ./...
