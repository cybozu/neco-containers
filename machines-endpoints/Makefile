SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: check-generate test

.PHONY: setup
setup:
	aqua i -l

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint:
	golangci-lint run -vvv ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run -vvv --fix ./...

.PHONY: test
test:
	$(MAKE) lint
	$(SUDO_GO) test -race -v ./...
