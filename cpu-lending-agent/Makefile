.PHONY: all
all: check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint:
	aqua install
	golangci-lint run

.PHONY: test
test:
	go vet ./...
	go test -race ./...

.PHONY: docker-build
docker-build:
	docker build -t ghcr.io/cybozu/cpu-lending-agent:dev .

.PHONY: e2e
e2e:
	$(MAKE) -C e2e start
	$(MAKE) -C e2e test
	$(MAKE) -C e2e stop

include ../tool.mk
