include ../tool.mk

.PHONY: lint
lint:
	aqua install
	golangci-lint run
	cd plugin/incluster-filter; golangci-lint run

.PHONY: test
test:
	$(MAKE) -C e2e setup
	$(MAKE) -C e2e start
	$(MAKE) -C e2e test
	$(MAKE) -C e2e stop

.PHONY: docker-build
docker-build:
	docker build -t ghcr.io/cybozu/ct-monitor:dev .
