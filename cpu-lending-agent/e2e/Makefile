KIND_CLUSTER_NAME := cpu-lending-agent
KIND_CONFIG := kind.yaml
IMAGE := ghcr.io/cybozu/cpu-lending-agent:dev

KIND ?= kind
KUBECTL ?= kubectl

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --config=$(KIND_CONFIG)
	$(MAKE) -C ../ docker-build
	$(KIND) load docker-image $(IMAGE) --name=$(KIND_CLUSTER_NAME)

.PHONY: test
test:
	env RUN_E2E=1 KUBECTL=$(KUBECTL) \
		go test -v -race . -ginkgo.v -ginkgo.fail-fast

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)
