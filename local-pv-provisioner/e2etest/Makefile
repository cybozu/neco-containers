# It is set by CI using the environment variable, use conditional assignment.
KUBERNETES_VERSION := 1.27.10
# https://github.com/kubernetes/minikube/releases
MINIKUBE_VERSION := v1.32.0

# Tools versions which are defined in go.mod
SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GINKGO_VERSION := $(shell awk '$$1 == "github.com/onsi/ginkgo/v2" {print $$2}' $(SELF_DIR)/../go.mod)

SHELL := /bin/bash
BINDIR := $(shell pwd)/bin
CURL := curl -sSLf
GINKGO := $(BINDIR)/ginkgo-$(GINKGO_VERSION)
GINKGO_FLAGS :=
MINIKUBE := $(BINDIR)/minikube
KUBECTL := $(BINDIR)/kubectl-$(KUBERNETES_VERSION)
MINIKUBE_HOME = $(BINDIR)
NODE_NAME := minikube-worker
SUDO := sudo
DD := dd
LOSETUP := losetup
LOOP_FILE := loop.img
LOOP_FILE2 := loop2.img
MINIKUBE_PROFILE := lpptest

export MINIKUBE_HOME

.PHONY: setup
setup:
	$(MAKE) $(GINKGO)
	$(MAKE) $(KUBECTL)
	$(MAKE) $(MINIKUBE)

.PHONY: test
test:
	$(MAKE) launch-cluster MINIKUBE_PROFILE=$(MINIKUBE_PROFILE)
	$(MAKE) create-loop-dev
	$(MAKE) launch-local-pv-provisioner
	$(MAKE) do-test

.PHONY: clean
clean:
	$(MINIKUBE) delete --all || true

$(BINDIR):
	mkdir -p $@

$(GINKGO): | $(BINDIR)
	GOBIN=$(BINDIR) go install github.com/onsi/ginkgo/v2/ginkgo@$(GINKGO_VERSION)
	mv $(BINDIR)/ginkgo $@

$(KUBECTL): | $(BINDIR)
	$(CURL) -o $@ https://storage.googleapis.com/kubernetes-release/release/v$(KUBERNETES_VERSION)/bin/linux/amd64/kubectl
	chmod a+x $@

$(MINIKUBE): | $(BINDIR)
	$(CURL) -o $@ https://github.com/kubernetes/minikube/releases/download/$(MINIKUBE_VERSION)/minikube-linux-amd64
	chmod a+x $@

.PHONY: launch-cluster
launch-cluster: MINIKUBE_PROFILE=
launch-cluster:
	# TODO: Is there any better way to verify whether k8s cluster is available or not?
	if $(MINIKUBE) profile $(MINIKUBE_PROFILE) |& grep "not found" > /dev/null; then \
		$(MINIKUBE) start \
			--kubernetes-version="v$(KUBERNETES_VERSION)" \
			--driver=kvm2 \
			--memory=8g \
			--cpus=4 \
			--extra-config=kubeadm.node-name=$(NODE_NAME) \
			--extra-config=kubelet.hostname-override=$(NODE_NAME) \
			-p $(MINIKUBE_PROFILE) ; \
	fi
	$(MINIKUBE) profile $(MINIKUBE_PROFILE)
	$(MAKE) image-build

.PHONY: create-loop-dev
create-loop-dev:
	$(MINIKUBE) ssh -- $(DD) if=/dev/zero of=loop0.img bs=1G seek=1 count=0
	$(MINIKUBE) ssh -- $(SUDO) $(LOSETUP) /dev/loop0 loop0.img || :
	$(MINIKUBE) ssh -- $(DD) if=/dev/zero of=loop1.img bs=1G seek=1 count=0
	$(MINIKUBE) ssh -- $(SUDO) $(LOSETUP) /dev/loop1 loop1.img || :
	$(MINIKUBE) ssh -- lsblk
	$(MINIKUBE) ssh -- $(SUDO) mkdir -p /dev/e2etest
	$(MINIKUBE) ssh -- $(SUDO) mknod /dev/e2etest/loop0 b 7 0 || :
	$(MINIKUBE) ssh -- $(SUDO) mknod /dev/e2etest/loop1 b 7 1 || :

.PHONY: image-build
image-build:
	eval $$($(MINIKUBE) docker-env); \
	$(MAKE) -C .. docker
	$(MINIKUBE) ssh -- docker images

.PHONY: launch-local-pv-provisioner
launch-local-pv-provisioner:
	$(KUBECTL) apply -f testdata/local-pv-provisioner.yaml

.PHONY: do-test
do-test: $(GINKGO)
	env \
	PATH=${PATH} \
	E2ETEST=1 \
	KUBECTL=$(KUBECTL) \
	MINIKUBE=$(MINIKUBE) \
	$(GINKGO) --fail-fast -v $(GINKGO_FLAGS) .
