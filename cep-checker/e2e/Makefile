E2ETEST_K8S_VERSION := 1.34.8
E2ETEST_K8S_IMAGE   := kindest/node:v1.34.8@sha256:02722c2dedddcfc00febf5d27fbeb9b7b2c14294c82109ff4a85d89ac9ba3256
CILIUM_VERSION      := 1.16.19

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := cep-checker
KIND_CONFIG := kind.yaml

CILIUM_CLI := $(BIN_DIR)/cilium

.PHONY: setup
setup: | $(BIN_DIR)
	aqua install
	cp $$(aqua which kind) $(KIND)
	cp $$(aqua which kubectl) $(KUBECTL)
	cp $$(aqua which cilium) $(CILIUM_CLI)

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=$(E2ETEST_K8S_IMAGE) --config=$(KIND_CONFIG)
	$(MAKE) -C ../ docker-build
	$(KIND) load docker-image ghcr.io/cybozu/cep-checker:dev --name=$(KIND_CLUSTER_NAME)

.PHONY: install-cilium
install-cilium:
	$(CILIUM_CLI) install --version $(CILIUM_VERSION) --wait

.PHONY: test
test:
	env RUN_E2E=1 \
		go test -v -race . -ginkgo.v -ginkgo.fail-fast

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)
	-docker image rm ghcr.io/cybozu/cep-checker:dev

$(BIN_DIR):
	mkdir -p $@
