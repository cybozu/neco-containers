.PHONY: all
all: check-generated test

.PHONY: setup
setup:
	aqua i -l

.PHONY: check-generated
check-generated:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint:
	aqua exec -- golangci-lint run ./... -vvv

.PHONY: lint-fix
lint-fix:
	aqua exec -- golangci-lint run ./... -vvv --fix

.PHONY: test
test:
	$(MAKE) lint
	go test -race -v ./...
