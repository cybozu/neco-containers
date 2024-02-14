.PHONY: all
all: check-generate test

.PHONY: setup
setup:
	$(MAKE) -C src setup

.PHONY: check-generate
check-generate:
	$(MAKE) -C src check-generate

.PHONY: test
test:
	$(MAKE) -C src test
