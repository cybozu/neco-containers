ARCH ?= amd64
OS ?= linux

E2ETEST_K8S_VERSION  := 1.33.4
KIND_VERSION := 0.30.0
CILIUM_VERSION := $(shell go list -m -f '{{.Version}}' github.com/cilium/cilium | cut -c2-)
CILIUM_CLI_VERSION := 0.18.8

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

CURL := curl -sSLf
KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := neco-server-exporter
KIND_CONFIG := testdata/kind.yaml

CILIUM_CLI := $(BIN_DIR)/cilium

IMAGE_TAG ?= ghcr.io/cybozu/neco-server-exporter:dev

.PHONY: setup
setup: $(KUBECTL) $(KIND) $(CILIUM_CLI)

.PHONY: prepare
prepare:
	sudo sysctl -w kernel.bpf_stats_enabled=1

.PHONY: wait-for-workloads
wait-for-workloads:
	$(KUBECTL) wait --for=condition=Available --all deployments --all-namespaces --timeout=1h
	$(KUBECTL) wait --for=condition=Ready --all pods --all-namespaces --timeout=1h

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=kindest/node:v$(E2ETEST_K8S_VERSION) --config=$(KIND_CONFIG)
	$(CILIUM_CLI) install --version $(CILIUM_VERSION) --wait

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: install
install:
	$(MAKE) -C ../ IMAGE_TAG=$(IMAGE_TAG) docker-build
	$(KIND) load docker-image $(IMAGE_TAG) --name=$(KIND_CLUSTER_NAME)
	$(KUBECTL) apply -f testdata/namespace.yaml
	$(KUBECTL) apply -f testdata/daemonset.yaml
	$(KUBECTL) apply -f testdata/service.yaml
	$(MAKE) --no-print-directory wait-for-workloads

.PHONY: uninstall
uninstall:
	$(KUBECTL) delete -f testdata/namespace.yaml
	-docker image rm $(IMAGE_TAG)
	-docker image prune -f

.PHONY: scrape
scrape:
	@PORT=$$($(KUBECTL) get service -n neco-server-exporter neco-server-exporter "-o=jsonpath={ .spec.ports[0].nodePort }"); \
	docker exec neco-server-exporter-control-plane curl -s http://localhost:$${PORT}/metrics

.PHONY: test
test:
	go test -v -race . -ginkgo.v -ginkgo.fail-fast

$(KIND): $(BIN_DIR)
	$(CURL) -o $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-$(OS)-$(ARCH)
	chmod a+x $(KIND)

$(KUBECTL): $(BIN_DIR)
	$(CURL) -o $(BIN_DIR)/kubectl https://dl.k8s.io/v$(E2ETEST_K8S_VERSION)/bin/$(OS)/$(ARCH)/kubectl && chmod a+x $(BIN_DIR)/kubectl

$(CILIUM_CLI): $(BIN_DIR)
	$(CURL) https://github.com/cilium/cilium-cli/releases/download/v$(CILIUM_CLI_VERSION)/cilium-linux-amd64.tar.gz | tar -xz -C $(BIN_DIR)
	chmod a+x $@

$(BIN_DIR):
	mkdir -p $@
