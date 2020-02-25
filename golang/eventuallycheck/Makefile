TARGET = eventuallycheck

all: build

setup:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u golang.org/x/lint/golint

mod:
	go mod tidy
	go mod vendor

build: mod $(TARGET)

$(TARGET):
	CGO_ENABLED=0 go build -o $@ ./cmd/$@

clean:
	-rm $(TARGET)

fmt:
	goimports -w $$(find . -type d -name 'vendor' -prune -o -type f -name '*.go' -print)

run: build
	./$(TARGET) ./testdata/*

test:
	test -z "$$(goimports -l $$(find . -type d -name 'vendor' -prune -o -type f -name '*.go' -print) | tee /dev/stderr)"
	test -z "$$(golint $$(go list ./... | grep -v '/vendor/') | tee /dev/stderr)"
	CGO_ENABLED=0 go test -v ./...

.PHONY: all setup mod build clean fmt run test
