SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: check-generate test

.PHONY: setup
setup:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: test
test:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	go vet ./...
	$(SUDO_GO) test -race -v ./...
