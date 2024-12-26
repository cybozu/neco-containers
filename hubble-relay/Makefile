IMAGE_TAG ?= ghcr.io/cybozu/hubble-relay:$(shell cat TAG)

.PHONY: build
build:
	docker build . --tag=$(IMAGE_TAG)

.PHONY: cilium-checkout
cilium-checkout:
	@$(MAKE) -C ../cilium checkout

.PHONY: cilium-test-e2e-setup
cilium-test-e2e-setup:
	@$(MAKE) -C ../cilium test-e2e-setup

.PHONY: cilium-test-e2e-upgrade-inotify
cilium-test-e2e-upgrade-inotify:
	@$(MAKE) -C ../cilium test-e2e-upgrade-inotify

.PHONY: test-e2e
test-e2e:
	@$(MAKE) -C ../cilium/e2e start-hubble-relay
	kubectl wait deployments -A --all --for condition=Available --timeout=300s
	kubectl wait pods -A --all --for condition=Ready --timeout=300s
	@$(MAKE) -C ../cilium/e2e stop
