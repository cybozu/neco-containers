SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: test

.PHONY: test
test:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	$(SUDO_GO) test -race -v ./...
