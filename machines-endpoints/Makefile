SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: test

.PHONY: test
test:
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | grep -v vendor/ | tee /dev/stderr)"
	ineffassign .
	go vet ./...
	$(SUDO_GO) test -race -v ./...
