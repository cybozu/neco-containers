IMAGE_TAG ?= ghcr.io/cybozu/neco-exporter:$(shell cat TAG)

.PHONY: check-generate
check-generate: goimports
	go mod tidy
	$(GOIMPORTS) -w -local github.com/cybozu/neco-containers/neco-exporter .
	git diff --exit-code --name-only

.PHONY build:
build:
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/neco-exporter .

.PHONY: e2e-prepare
e2e-prepare:
	$(MAKE) -C e2e prepare

.PHONY: test
test:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	golangci-lint run

	$(MAKE) -C e2e setup
	$(MAKE) -C e2e start
	$(MAKE) -C e2e install
	$(MAKE) -C e2e test
	$(MAKE) -C e2e stop

.PHONY: docker-build
docker-build:
	docker build . --tag=$(IMAGE_TAG)

.PHONY: sanity-check
sanity-check:
	docker run --rm $(IMAGE_TAG) --help

include ../tool.mk
