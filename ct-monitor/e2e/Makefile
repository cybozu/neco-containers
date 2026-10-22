E2ETEST_K8S_VERSION      := 1.35.8
E2ETEST_K8S_IMAGE        := kindest/node:v1.35.8@sha256:07b2536e30b803ed61d1677a79df6115f798ce64c80f9e22f6ed45afd09323c0
CERT_MANAGER_VERSION     := 1.21.2
CERT_MANAGER_CRDS_SHA256 := 262fef78478492cd35b73a1b227106b7802f9c056361d1f561d04b32d751337f

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
