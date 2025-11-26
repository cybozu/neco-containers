GOARCH ?= $(shell go env GOARCH)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

E2ETEST_K8S_VERSION  := 1.34.0
KIND_VERSION := 0.30.0
CILIUM_VERSION := 1.16.12
CILIUM_CLI_VERSION := 0.18.8

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

CURL := curl -sSLf
KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := cep-checker
KIND_CONFIG := kind.yaml

CILIUM_CLI := $(BIN_DIR)/cilium

.PHONY: setup
setup: $(KUBECTL) $(KIND) $(CILIUM_CLI)

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=kindest/node:v$(E2ETEST_K8S_VERSION) --config=$(KIND_CONFIG)
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

$(KIND): $(BIN_DIR)
	$(CURL) -o $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-$(OS)-$(GOARCH)
	chmod a+x $(KIND)

$(KUBECTL): $(BIN_DIR)
	$(CURL) -o $(BIN_DIR)/kubectl https://dl.k8s.io/v$(E2ETEST_K8S_VERSION)/bin/$(OS)/$(GOARCH)/kubectl && chmod a+x $(BIN_DIR)/kubectl

$(CILIUM_CLI): $(BIN_DIR)
	$(CURL) https://github.com/cilium/cilium-cli/releases/download/v$(CILIUM_CLI_VERSION)/cilium-$(OS)-$(GOARCH).tar.gz | tar -xz -C $(BIN_DIR)
	chmod a+x $@

$(BIN_DIR):
	mkdir -p $@
