E2ETEST_K8S_VERSION := 1.35.8
E2ETEST_K8S_IMAGE   := kindest/node:v1.35.8@sha256:07b2536e30b803ed61d1677a79df6115f798ce64c80f9e22f6ed45afd09323c0

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := squid-exporter-cluster

.PHONY: setup
setup: | $(BIN_DIR)
	aqua install
	cp $$(aqua which kind) $(KIND)
	cp $$(aqua which kubectl) $(KUBECTL)

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=$(E2ETEST_K8S_IMAGE) --config=kind-config.yaml --wait 1m
	$(MAKE) -C ../ docker-build
	$(KIND) load docker-image ghcr.io/cybozu/squid-exporter:dev --name=$(KIND_CLUSTER_NAME)

.PHONY: test
test:
	env RUN_E2E=1 \
		go test -v -race . -ginkgo.v -ginkgo.fail-fast

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)
	-docker image rm ghcr.io/cybozu/squid-exporter:dev

$(BIN_DIR):
	mkdir -p $@
