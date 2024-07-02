GO=go
GO_VET_OPTS=-v
GO_TEST_OPTS=-v -race

GO_FMT=gofmt
GO_FMT_OPTS=-s -l

DOCKER=docker
DOCKER_FILE=./Dockerfile
DOCKER_REPO=ghcr.io/cybozu
DOCKER_TAG=dev
DOCKER_CONTEXT=.

CUSTOM_CHECKER=custom-checker
STATIC_CHECK=staticcheck

$(CUSTOM_CHECKER):
	go install github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest

$(STATIC_CHECK):
	go install honnef.co/go/tools/cmd/staticcheck@latest


.PHONY: fmt
fmt:
	$(GO_FMT) $(GO_FMT_OPTS) .

.PHONY: vet
vet:
	$(GO) vet $(GO_VET_OPTS) ./...

.PHONY: mod
mod:
	$(GO) mod tidy

.PHONY: check-diff
check-diff: mod fmt
	git diff --exit-code --name-only

.PHONY: test
test: vet $(STATIC_CHECK) $(CUSTOM_CHECKER)
	staticcheck ./...
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	$(GO) test $(GO_TEST_OPTS) ./...

.PHONY: build
build:
	$(GO) build $(GO_BUILD_OPT) -o ./bin/tcp-keepalive .

.PHONY: docker
docker:
	$(DOCKER) image build -f $(DOCKER_FILE) -t $(DOCKER_REPO)/tcp-keepalive:$(DOCKER_TAG) $(DOCKER_CONTEXT)

.PHONY: clean
clean:
	-$(GO) clean
	-rm $(RM_OPTS) ./bin/*

.PHONY: all
all: check-diff test build

.DEFAULT_GOAL=all
