E2ETEST_K8S_VERSION := 1.35.1
E2ETEST_K8S_IMAGE   := kindest/node:v1.35.1@sha256:05d7bcdefbda08b4e038f644c4df690cdac3fba8b06f8289f30e10026720a1ab
CILIUM_VERSION := $(shell go list -m -f '{{.Version}}' github.com/cilium/cilium | cut -c2-)

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

KUBECTL := $(BIN_DIR)/kubectl

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := neco-exporter
KIND_CONFIG := testdata/kind.yaml

CILIUM_CLI := $(BIN_DIR)/cilium

YQ := $(BIN_DIR)/yq

IMAGE_TAG ?= ghcr.io/cybozu/neco-exporter:dev

# CSI_ADDONS_VERSION should be updated with csi-addons-controller version. See `csi-addons-controller/BRANCH` .
CSI_ADDONS_VERSION := v0.14.0

.PHONY: setup
setup: | $(BIN_DIR)
	aqua install
	cp $$(aqua which kind) $(KIND)
	cp $$(aqua which kubectl) $(KUBECTL)
	cp $$(aqua which cilium) $(CILIUM_CLI)
	cp $$(aqua which yq) $(YQ)

.PHONY: prepare
prepare:
	sudo sysctl -w kernel.bpf_stats_enabled=1

.PHONY: wait-for-workloads
wait-for-workloads:
	$(KUBECTL) wait --for=condition=Available --all deployments --all-namespaces --timeout=1h
	$(KUBECTL) wait --for=condition=Ready --all pods --all-namespaces --timeout=1h

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --image=$(E2ETEST_K8S_IMAGE) --config=$(KIND_CONFIG)
	$(CILIUM_CLI) install --version $(CILIUM_VERSION) --wait

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: install-pilot
install-pilot:
	$(KUBECTL) apply -f testdata/pilot.yaml
	$(MAKE) --no-print-directory wait-for-workloads

	PODNAME=$$($(KUBECTL) get po -l app=pilot -o name | cut -d'/' -f2); \
	$(KUBECTL) cp testdata/onboard $${PODNAME}:/tmp/; \
	$(KUBECTL) exec $${PODNAME} -- chmod +x /tmp/onboard; \
	$(KUBECTL) exec $${PODNAME} -- /tmp/onboard; \
	$(KUBECTL) cp $(KUBECTL) $${PODNAME}:/tmp/; \
	$(KUBECTL) exec $${PODNAME} -- chmod +x /tmp/kubectl

.PHONY: install
install: install-pilot
	$(MAKE) -C ../ IMAGE_TAG=$(IMAGE_TAG) docker-build
	$(KIND) load docker-image $(IMAGE_TAG) --name=$(KIND_CLUSTER_NAME)
	$(KUBECTL) apply -f testdata/csi-addons-crds.yaml
	$(KUBECTL) apply -f testdata/namespace.yaml
	$(KUBECTL) apply -f testdata/serviceaccount.yaml
	$(KUBECTL) apply -f testdata/role.yaml
	$(KUBECTL) apply -f testdata/rolebinding.yaml
	$(KUBECTL) apply -f testdata/daemonset.yaml
	$(KUBECTL) apply -f testdata/deployment.yaml
	$(KUBECTL) apply -f testdata/service.yaml
	$(MAKE) --no-print-directory wait-for-workloads

# Extract only the NetworkFence CRD; the neco-exporter only watches NetworkFence resources.
.PHONY: update-csi-addons-crd
update-csi-addons-crd:
	curl -sSL https://raw.githubusercontent.com/csi-addons/kubernetes-csi-addons/$(CSI_ADDONS_VERSION)/deploy/controller/crds.yaml \
		| $(YQ) 'select(.metadata.name == "networkfences.csiaddons.openshift.io")' \
		> testdata/csi-addons-crds.yaml

.PHONY: uninstall
uninstall:
	$(KUBECTL) delete -f testdata/namespace.yaml
	-docker image rm $(IMAGE_TAG)

.PHONY: pilot
pilot:
	@PODNAME=$$($(KUBECTL) get po -l app=pilot -o name | cut -d'/' -f2); \
	$(KUBECTL) exec -it $${PODNAME} -- bash

.PHONY: scrape-cluster
scrape-cluster:
	$(KUBECTL) exec deploy/pilot -it -- curl -s http://neco-cluster-exporter.neco-exporter.svc/metrics

.PHONY: scrape-node
scrape-node:
	$(KUBECTL) exec deploy/pilot -it -- curl -s http://neco-node-exporter.neco-exporter.svc/metrics

.PHONY: test
test:
	go test -v -race . -ginkgo.v -ginkgo.fail-fast

$(BIN_DIR):
	mkdir -p $@
