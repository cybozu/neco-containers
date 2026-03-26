include ../tool.mk

SERVER_IMAGE ?= ghcr.io/cybozu/spiffe-test-server:$(shell cat TAG)
CLIENT_IMAGE ?= ghcr.io/cybozu/spiffe-test-client:$(shell cat TAG)

GO_IMPORTS_OPTS := -w -local github.com/cybozu/neco-containers/spiffe-test

# Protocol Buffers
.PHONY: proto
proto: buf
	$(BUF) generate

.PHONY: proto-lint
proto-lint: buf
	$(BUF) lint

# Testing and linting
.PHONY: vet
vet:
	go vet -v ./...

.PHONY: mod
mod:
	go mod tidy

.PHONY: fmt
fmt: gofumpt goimports
	$(GOFUMPT) -l -w .
	$(GOIMPORTS) $(GO_IMPORTS_OPTS) .

.PHONY: test
test: staticcheck
	$(STATICCHECK) ./...
	go test -v -race ./...

.PHONY: check-diff
check-diff: mod fmt proto-lint proto
	git diff --exit-code --name-only

.PHONY: check
check: vet check-diff test

# Build Go binaries
.PHONY: build
build:
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/client ./cmd/client

# Build Docker images
.PHONY: docker-build
docker-build:
	docker build --target server -t $(SERVER_IMAGE) .
	docker build --target client -t $(CLIENT_IMAGE) .

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/

.DEFAULT_GOAL := check
