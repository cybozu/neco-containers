SUDO=sudo
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOFLAGS = -mod=vendor
export GOFLAGS

all: build

test:
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
	test -z "$$(golint $$(go list ./... | grep -v /vendor/) | grep -v '/mtest/.*: should not use dot imports' | tee /dev/stderr)"
	test -z "$$(nilerr $$(go list ./... | grep -v /vendor/) 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | grep -v vendor/ | tee /dev/stderr)"
	ineffassign .
	go test -race -v ./...
	go vet ./...

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/ingress-watcher main.go

clean:
	rm -rf bin

.PHONY: all test build clean
