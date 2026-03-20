GOARCH ?= $(shell go env GOARCH)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

E2ETEST_K8S_VERSION := 1.34.3
KIND_VERSION := 0.30.0
CERT_MANAGER_VERSION := 1.19.4

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

CURL := curl -sSLf
KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := ct-monitor

.PHONY: setup
setup: bin_dir kubectl kind

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=kindest/node:v$(E2ETEST_K8S_VERSION) --wait 1m
	$(KUBECTL) apply -f https://github.com/cert-manager/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.crds.yaml
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
	chmod a+x $(KIND)

.PHONY: kubectl
kubectl: $(KUBECTL)
$(KUBECTL): $(BIN_DIR)
	$(CURL) -o $(BIN_DIR)/kubectl https://dl.k8s.io/v$(E2ETEST_K8S_VERSION)/bin/$(OS)/$(GOARCH)/kubectl && chmod a+x $(BIN_DIR)/kubectl

.PHONY: bin_dir
$(BIN_DIR):
	mkdir -p $(BIN_DIR)
