SUDO ?= sudo
IMAGE_TAG ?= ghcr.io/cybozu/cilium:$(shell cat ../TAG)

.PHONY: setup
setup:
	mkdir -p bin
	curl -sL -o- https://github.com/cilium/cilium-cli/releases/latest/download/cilium-linux-amd64.tar.gz | tar -xz -C bin/
	helm repo add cilium https://helm.cilium.io/
	helm repo update

# If connectivity test fails with "too many open files", run the follwing command
# https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files
.PHONY: upgrade-inotify
upgrade-inotify:
	$(SUDO) sysctl fs.inotify.max_user_watches=524288
	$(SUDO) sysctl fs.inotify.max_user_instances=512

.PHONY: start
start:
	kind create cluster --config cluster.yaml
	kind load docker-image $(IMAGE_TAG)
	helm install cilium cilium/cilium --version $(shell cat ../TAG | cut -d'.' -f1-3) \
		--namespace kube-system \
		--values values.yaml \
		--set image.repository=$(shell echo $(IMAGE_TAG) | cut -d':' -f1) \
		--set image.tag=$(shell echo $(IMAGE_TAG) | cut -d':' -f2)
	kubectl rollout status daemonsets -n kube-system cilium --timeout=120s
	kubectl wait deployments -A --all --for condition=Available --timeout=120s

.PHONY: test
test:
	bin/cilium connectivity test

.PHONY: stop
stop:
	kind delete cluster
