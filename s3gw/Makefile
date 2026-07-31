SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	cd tools && go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint:
	go tool -modfile=tools/go.mod golangci-lint run ./... -vvv

.PHONY: lint-fix
lint-fix:
	go tool -modfile=tools/go.mod golangci-lint run ./... -vvv --fix

.PHONY: test
test:
	$(MAKE) lint
	$(SUDO_GO) test -race -v ./...
