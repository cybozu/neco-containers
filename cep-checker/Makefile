
include ../tool.mk

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o cep-checker .

.PHONY: test
test:
	$(MAKE) -C e2e setup
	$(MAKE) -C e2e start
	$(MAKE) -C e2e install-cilium
	$(MAKE) -C e2e test
	$(MAKE) -C e2e stop

.PHONY: docker-build
docker-build:
	docker build -t ghcr.io/cybozu/cep-checker:dev .

.PHONY: check-generate
check-generate: goimports
	go mod tidy
	$(GOIMPORTS) -w -local github.com/cybozu/neco-containers/cep-checker .
	git diff --exit-code --name-only
