include ../tool.mk

.PHONY: test
test:
	$(MAKE) -C e2e setup
	$(MAKE) -C e2e start
	$(MAKE) -C e2e test
	$(MAKE) -C e2e stop

.PHONY: docker-build
docker-build:
	docker build -t ghcr.io/cybozu/ct-monitor:dev .
