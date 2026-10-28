E2ETEST_K8S_VERSION := 1.35.8
E2ETEST_K8S_IMAGE   := kindest/node:v1.35.8@sha256:07b2536e30b803ed61d1677a79df6115f798ce64c80f9e22f6ed45afd09323c0
CILIUM_VERSION      := 1.17.12

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
