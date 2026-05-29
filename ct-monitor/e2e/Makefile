E2ETEST_K8S_VERSION      := 1.34.3
E2ETEST_K8S_IMAGE        := kindest/node:v1.34.3@sha256:08497ee19eace7b4b5348db5c6a1591d7752b164530a36f855cb0f2bdcbadd48
CERT_MANAGER_VERSION     := 1.19.4
CERT_MANAGER_CRDS_SHA256 := 0e09c26d64fbdb56f2ea95b5b2350991980a184a9ce9a09d3a9a207594ce38b0

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

CURL := curl -sSLf
KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := ct-monitor

.PHONY: setup
setup: | $(BIN_DIR)
	aqua install
	cp $$(aqua which kind) $(KIND)
	cp $$(aqua which kubectl) $(KUBECTL)

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=$(E2ETEST_K8S_IMAGE) --wait 1m
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

$(BIN_DIR):
	mkdir -p $@
