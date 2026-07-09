E2ETEST_K8S_VERSION      := 1.34.8
E2ETEST_K8S_IMAGE        := kindest/node:v1.34.8@sha256:02722c2dedddcfc00febf5d27fbeb9b7b2c14294c82109ff4a85d89ac9ba3256
CERT_MANAGER_VERSION     := 1.21.0
CERT_MANAGER_CRDS_SHA256 := cc046deb0ef4aaa8237f1ae448ef463446905bfc310cba88f25fc614389f84fd

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
