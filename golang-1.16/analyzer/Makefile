TARGET = custom-checker eventuallycheck restrictpkg

SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: test build

.PHONY: setup
setup:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u golang.org/x/lint/golint

.PHONY: build
build: $(TARGET)

$(TARGET):
	CGO_ENABLED=0 go build ./cmd/$@

.PHONY: clean
clean:
	-rm $(TARGET)

.PHONY: test
test:
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	$(SUDO_GO) test -race -v ./...
