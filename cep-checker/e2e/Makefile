GOARCH ?= $(shell go env GOARCH)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

E2ETEST_K8S_VERSION   := 1.34.3
E2ETEST_K8S_IMAGE     := kindest/node:v1.34.3@sha256:babda82416d417f720a4d6dbd35deec5263af2a6c164c81c08cde0044c2b9f78
KIND_VERSION          := 0.30.0
CILIUM_VERSION        := 1.16.19
CILIUM_CLI_VERSION    := 0.19.2

KUBECTL_SHA256_linux_amd64    := ab60ca5f0fd60c1eb81b52909e67060e3ba0bd27e55a8ac147cbc2172ff14212
KIND_SHA256_linux_amd64       := 517ab7fc89ddeed5fa65abf71530d90648d9638ef0c4cde22c2c11f8097b8889
CILIUM_CLI_SHA256_linux_amd64 := aa25cd2542614051573008ed18adb7d697fd78e4b0345db83d1650cdaf4492de

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

$(KIND): $(BIN_DIR)
	$(CURL) -o $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-$(OS)-$(GOARCH)
	$(if $(KIND_SHA256_$(OS)_$(GOARCH)),echo "$(KIND_SHA256_$(OS)_$(GOARCH))  $(KIND)" | sha256sum -c)
	chmod a+x $(KIND)

$(KUBECTL): $(BIN_DIR)
	$(CURL) -o $(BIN_DIR)/kubectl https://dl.k8s.io/v$(E2ETEST_K8S_VERSION)/bin/$(OS)/$(GOARCH)/kubectl
	$(if $(KUBECTL_SHA256_$(OS)_$(GOARCH)),echo "$(KUBECTL_SHA256_$(OS)_$(GOARCH))  $(BIN_DIR)/kubectl" | sha256sum -c)
	chmod a+x $(BIN_DIR)/kubectl

$(CILIUM_CLI): $(BIN_DIR)
	$(CURL) -o $(BIN_DIR)/cilium-cli.tar.gz https://github.com/cilium/cilium-cli/releases/download/v$(CILIUM_CLI_VERSION)/cilium-$(OS)-$(GOARCH).tar.gz
	$(if $(CILIUM_CLI_SHA256_$(OS)_$(GOARCH)),echo "$(CILIUM_CLI_SHA256_$(OS)_$(GOARCH))  $(BIN_DIR)/cilium-cli.tar.gz" | sha256sum -c)
	tar -xz -C $(BIN_DIR) -f $(BIN_DIR)/cilium-cli.tar.gz
	rm $(BIN_DIR)/cilium-cli.tar.gz
	chmod a+x $@

$(BIN_DIR):
	mkdir -p $@
