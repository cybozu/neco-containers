ARCH ?= amd64
OS ?= linux

E2ETEST_K8S_VERSION  := 1.28.7
KIND_VERSION := 0.23.0
PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

CURL := curl -sSLf
KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := daemonset-updater
KIND_CONFIG := kind.yaml

.PHONY: setup
setup: $(KUBECTL) $(KIND)

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=kindest/node:v$(E2ETEST_K8S_VERSION) --config=$(KIND_CONFIG)
	$(MAKE) -C ../ docker-build
	$(KIND) load docker-image ghcr.io/cybozu/daemonset-updater:dev --name=$(KIND_CLUSTER_NAME)
	$(KUBECTL) label nodes --overwrite $(KIND_CLUSTER_NAME)-control-plane cke.cybozu.com/rack=1
	$(KUBECTL) label nodes --overwrite $(KIND_CLUSTER_NAME)-worker cke.cybozu.com/rack=1
	$(KUBECTL) label nodes --overwrite $(KIND_CLUSTER_NAME)-worker2 cke.cybozu.com/rack=2
	$(KUBECTL) apply -f daemonset.yaml
	sleep 3
	$(KUBECTL) patch daemonset testhttpd --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"ghcr.io/cybozu/testhttpd:0.2.5"}]'



.PHONY: test
test:
	env RUN_E2E=1 \
		go test -v -race . -ginkgo.v -ginkgo.fail-fast

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)
	-docker image rm ghcr.io/cybozu/daemonset-updater:dev
	-docker image prune -f

$(KIND): $(BIN_DIR)
	$(CURL) -o $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-$(OS)-$(ARCH)
	chmod a+x $(KIND)

$(KUBECTL): $(BIN_DIR)
	$(CURL) -o $(BIN_DIR)/kubectl https://storage.googleapis.com/kubernetes-release/release/v$(E2ETEST_K8S_VERSION)/bin/$(OS)/$(ARCH)/kubectl && chmod a+x $(BIN_DIR)/kubectl

$(BIN_DIR):
	mkdir -p $@
