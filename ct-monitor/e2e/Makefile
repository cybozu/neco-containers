GOARCH ?= $(shell go env GOARCH)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

E2ETEST_K8S_VERSION := 1.34.3
KIND_VERSION := 0.30.0
CERT_MANAGER_VERSION := 1.19.4
CERT_MANAGER_CRDS_SHA256 := 0e09c26d64fbdb56f2ea95b5b2350991980a184a9ce9a09d3a9a207594ce38b0

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

CURL := curl -sSLf
KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := ct-monitor

KIND_SHA256_linux_amd64 := 517ab7fc89ddeed5fa65abf71530d90648d9638ef0c4cde22c2c11f8097b8889
KIND_SHA256_linux_arm64 := 7ea2de9d2d190022ed4a8a4e3ac0636c8a455e460b9a13ccf19f15d07f4f00eb
KIND_SHA256_darwin_amd64 := 4f0b6e3b88bdc66d922c08469f05ef507d4903dd236e6319199bb9c868eed274
KIND_SHA256_darwin_arm64 := ceaf40df1d1551c481fb50e3deb5c3deecad5fd599df5469626b70ddf52a1518
KIND_SHA256 := $(KIND_SHA256_$(OS)_$(GOARCH))

KUBECTL_SHA256_linux_amd64 := ab60ca5f0fd60c1eb81b52909e67060e3ba0bd27e55a8ac147cbc2172ff14212
KUBECTL_SHA256_linux_arm64 := 46913a7aa0327f6cc2e1cc2775d53c4a2af5e52f7fd8dacbfbfd098e757f19e9
KUBECTL_SHA256_darwin_amd64 := 657afbd0e653c4ce3af1b5a645a4eaba282cf8eb2bcda7191ff60866e50e4d7f
KUBECTL_SHA256_darwin_arm64 := e51367d2107d605f4edd7c2fb25897b0c0695a7de1a9f9d04cd6c9356b890b14
KUBECTL_SHA256 := $(KUBECTL_SHA256_$(OS)_$(GOARCH))

.PHONY: setup
setup: bin_dir kubectl kind

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=kindest/node:v$(E2ETEST_K8S_VERSION) --wait 1m
	$(CURL) -o $(BIN_DIR)/cert-manager.crds.yaml https://github.com/cert-manager/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.crds.yaml
	echo "$(CERT_MANAGER_CRDS_SHA256)  $(BIN_DIR)/cert-manager.crds.yaml" | sha256sum --check
	$(KUBECTL) apply -f $(BIN_DIR)/cert-manager.crds.yaml
	$(MAKE) -C ../ docker-build
	$(KIND) load docker-image ghcr.io/cybozu/ct-monitor:dev --name=$(KIND_CLUSTER_NAME)

.PHONY: test
test:
	env RUN_E2E=1 \
		go test -v -race . -ginkgo.v -ginkgo.fail-fast

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)
	-docker image rm ghcr.io/cybozu/ct-monitor:dev

.PHONY: kind
kind: $(KIND)
$(KIND): $(BIN_DIR)
	$(CURL) -o $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-$(OS)-$(GOARCH)
	echo "$(KIND_SHA256)  $(KIND)" | sha256sum --check
	chmod a+x $(KIND)

.PHONY: kubectl
kubectl: $(KUBECTL)
$(KUBECTL): $(BIN_DIR)
	$(CURL) -o $(KUBECTL) https://dl.k8s.io/v$(E2ETEST_K8S_VERSION)/bin/$(OS)/$(GOARCH)/kubectl
	echo "$(KUBECTL_SHA256)  $(KUBECTL)" | sha256sum --check
	chmod a+x $(KUBECTL)

.PHONY: bin_dir
$(BIN_DIR):
	mkdir -p $(BIN_DIR)
