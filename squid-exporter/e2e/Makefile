E2ETEST_K8S_VERSION := 1.34.0
E2ETEST_K8S_IMAGE   := kindest/node:v1.34.0@sha256:7416a61b42b1662ca6ca89f02028ac133a309a2a30ba309614e8ec94d976dc5a

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
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=$(E2ETEST_K8S_IMAGE) --wait 1m
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
