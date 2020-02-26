TARGET = custom-checker eventuallycheck restrictpkg
SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)


all: build

setup:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u golang.org/x/lint/golint

mod:
	go mod tidy
	go mod vendor

build: mod $(TARGET)

$(TARGET):
	CGO_ENABLED=0 go build ./cmd/$@

clean:
	-rm $(TARGET)

test:
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | grep -v '^testdata' | tee /dev/stderr)"
	test -z "$$(golint $$(go list ./... | grep -v /vendor/) | tee /dev/stderr)"
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	ineffassign .
	go vet ./...
	$(SUDO_GO) test -race -v ./...

.PHONY: all setup mod build clean test
