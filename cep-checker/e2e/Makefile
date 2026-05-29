GOARCH ?= $(shell go env GOARCH)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

E2ETEST_K8S_VERSION   := 1.34.3
E2ETEST_K8S_IMAGE     := kindest/node:v1.34.3@sha256:08497ee19eace7b4b5348db5c6a1591d7752b164530a36f855cb0f2bdcbadd48
KIND_VERSION          := 0.30.0
CILIUM_VERSION        := 1.16.19
CILIUM_CLI_VERSION    := 0.19.2

KIND_SHA256_linux_amd64       := 517ab7fc89ddeed5fa65abf71530d90648d9638ef0c4cde22c2c11f8097b8889
KIND_SHA256_linux_arm64       := 7ea2de9d2d190022ed4a8a4e3ac0636c8a455e460b9a13ccf19f15d07f4f00eb
KIND_SHA256_darwin_amd64      := 4f0b6e3b88bdc66d922c08469f05ef507d4903dd236e6319199bb9c868eed274
KIND_SHA256_darwin_arm64      := ceaf40df1d1551c481fb50e3deb5c3deecad5fd599df5469626b70ddf52a1518
KIND_SHA256                   := $(KIND_SHA256_$(OS)_$(GOARCH))

KUBECTL_SHA256_linux_amd64    := ab60ca5f0fd60c1eb81b52909e67060e3ba0bd27e55a8ac147cbc2172ff14212
KUBECTL_SHA256_linux_arm64    := 46913a7aa0327f6cc2e1cc2775d53c4a2af5e52f7fd8dacbfbfd098e757f19e9
KUBECTL_SHA256_darwin_amd64   := 657afbd0e653c4ce3af1b5a645a4eaba282cf8eb2bcda7191ff60866e50e4d7f
KUBECTL_SHA256_darwin_arm64   := e51367d2107d605f4edd7c2fb25897b0c0695a7de1a9f9d04cd6c9356b890b14
KUBECTL_SHA256                := $(KUBECTL_SHA256_$(OS)_$(GOARCH))

CILIUM_CLI_SHA256_linux_amd64  := aa25cd2542614051573008ed18adb7d697fd78e4b0345db83d1650cdaf4492de
CILIUM_CLI_SHA256_linux_arm64  := 205ad0eba18a105e4542a5ff41fea74929f8cb4af831d35c0fac591bfe48bbf1
CILIUM_CLI_SHA256_darwin_amd64 := 8b8c694129d6f9cceac0543af3934cf30a1de828a72c18b68d4bc7fc668c8c35
CILIUM_CLI_SHA256_darwin_arm64 := c1254bd1f94711804a1276f2f2f56297a615a8a6fa4f6ba077ea895e997e16bb
CILIUM_CLI_SHA256              := $(CILIUM_CLI_SHA256_$(OS)_$(GOARCH))

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
	echo "$(KIND_SHA256)  $(KIND)" | sha256sum --check
	chmod a+x $(KIND)

$(KUBECTL): $(BIN_DIR)
	$(CURL) -o $(KUBECTL) https://dl.k8s.io/v$(E2ETEST_K8S_VERSION)/bin/$(OS)/$(GOARCH)/kubectl
	echo "$(KUBECTL_SHA256)  $(KUBECTL)" | sha256sum --check
	chmod a+x $(KUBECTL)

$(CILIUM_CLI): $(BIN_DIR)
	$(CURL) -o $(BIN_DIR)/cilium-cli.tar.gz https://github.com/cilium/cilium-cli/releases/download/v$(CILIUM_CLI_VERSION)/cilium-$(OS)-$(GOARCH).tar.gz
	echo "$(CILIUM_CLI_SHA256)  $(BIN_DIR)/cilium-cli.tar.gz" | sha256sum --check
	tar -xz -C $(BIN_DIR) -f $(BIN_DIR)/cilium-cli.tar.gz
	rm $(BIN_DIR)/cilium-cli.tar.gz
	chmod a+x $@

$(BIN_DIR):
	mkdir -p $@
