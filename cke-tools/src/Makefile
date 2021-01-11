SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: setup check-uncommitted test

.PHONY: check-uncommitted
check-uncommitted:
	go mod tidy
	git diff --exit-code --name-only

# Test python scripts and go sources
.PHONY: test
test:
	$(MAKE) -C scripts setup test
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	$(SUDO_GO) test -race -v ./...


.PHONY: setup
setup:
	$(MAKE) -C scripts setup
