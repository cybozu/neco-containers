SUDO ?= sudo
CILIUM_DIR := ../src/cilium

CILIUM_AGENT_IMAGE_TAG ?= ghcr.io/cybozu/cilium:$(shell cat ../TAG)
HUBBLE_RELAY_IMAGE_TAG ?= ghcr.io/cybozu/hubble-relay:$(shell cat ../../hubble-relay/TAG)

.PHONY: setup
setup:
	aqua install
	helm repo add cilium https://helm.cilium.io/
	helm repo update

.PHONY: format
format:
	yq -i values.yaml

# If connectivity test fails with "too many open files", run the follwing command
# https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files
.PHONY: upgrade-inotify
upgrade-inotify:
	$(SUDO) sysctl fs.inotify.max_user_watches=524288
	$(SUDO) sysctl fs.inotify.max_user_instances=512

.PHONY: start
start:
	cd $(CILIUM_DIR); ./contrib/scripts/kind.sh --xdp "" 3 "" "" "none" "ipv4"
	kind load docker-image $(CILIUM_AGENT_IMAGE_TAG)
	cilium install --wait \
		--chart-directory=$(CILIUM_DIR)/install/kubernetes/cilium \
		--values values.yaml \
		--set image.repository=$(shell echo $(CILIUM_AGENT_IMAGE_TAG) | cut -d':' -f1) \
		--set image.tag=$(shell echo $(CILIUM_AGENT_IMAGE_TAG) | cut -d':' -f2)

.PHONY: start-hubble-relay
start-hubble-relay:
	cd $(CILIUM_DIR); ./contrib/scripts/kind.sh --xdp "" 3 "" "" "none" "ipv4"
	CILIUM_AGENT_IMAGE_TAG=quay.io/cilium/cilium:v$(shell echo $(HUBBLE_RELAY_IMAGE_TAG) | cut -d':' -f2 | cut -d'.' -f1-3); \
	docker pull $${CILIUM_AGENT_IMAGE_TAG}; \
	kind load docker-image $${CILIUM_AGENT_IMAGE_TAG}
	kind load docker-image $(HUBBLE_RELAY_IMAGE_TAG)
	cilium install --wait \
		--chart-directory=$(CILIUM_DIR)/install/kubernetes/cilium \
		--values values.yaml \
		--set hubble.relay.image.repository=$(shell echo $(HUBBLE_RELAY_IMAGE_TAG) | cut -d':' -f1) \
		--set hubble.relay.image.tag=$(shell echo $(HUBBLE_RELAY_IMAGE_TAG) | cut -d':' -f2) \
		--set hubble.relay.image.pullPolicy=Never \
		--set hubble.relay.securityContext.runAsUser=10000 \
		--set hubble.relay.securityContext.runAsGroup=10000

# check-log-errors may be disabled in CI to accomodate with minor known issues.
#
# However, it ignores other valuable error messages, so please check them manually
# when upgrading Cilium.
.PHONY: test
test:
	cilium connectivity test \
		--external-ip 8.8.8.8 \
		--external-other-ip 8.8.4.4 \
		--external-cidr 8.8.0.0/16 \
		--test=!to-fqdn,!policy-local-cluster-egress,!check-log-errors

.PHONY: stop
stop:
	kind delete cluster
