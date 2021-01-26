CONTROLLER_RUNTIME_VERSION := $(shell awk '/sigs\.k8s\.io\/controller-runtime/ {print substr($$2, 2)}' go.mod)
CONTROLLER_TOOLS_VERSION = 0.4.1
KUSTOMIZE_VERSION = 3.8.7
CALICO_VERSION = 3.16
CONTOUR_VERSION = 1.9.0
ARGOCD_VERSION = 1.7.7
GRAFANA_OPERATOR_VERSION = 3.7.0

# Set the shell used to bash for better error handling.
SHELL = /bin/bash
.SHELLFLAGS = -e -o pipefail -c

.PHONY: all
all: build

.PHONY: crds
crds:
	mkdir -p config/crd/bases
	curl -fsL -o config/crd/bases/calico.yaml https://docs.projectcalico.org/v$(CALICO_VERSION)/manifests/crds.yaml
	curl -fsL -o config/crd/bases/contour.yaml https://raw.githubusercontent.com/projectcontour/contour/v$(CONTOUR_VERSION)/examples/contour/01-crds.yaml
	curl -fsL -o config/crd/bases/application.yaml https://raw.githubusercontent.com/argoproj/argo-cd/v$(ARGOCD_VERSION)/manifests/crds/application-crd.yaml
	curl -fsL -o config/crd/bases/grafanadashboard.yaml https://github.com/integr8ly/grafana-operator/raw/v$(GRAFANA_OPERATOR_VERSION)/deploy/crds/GrafanaDashboard.yaml

# Run tests, and set up envtest if not done already.
ENVTEST_ASSETS_DIR := testbin
ENVTEST_SCRIPT_URL := https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v$(CONTROLLER_RUNTIME_VERSION)/hack/setup-envtest.sh
.PHONY: test
test: crds simple-test
ifeq (,$(wildcard $(ENVTEST_ASSETS_DIR)/setup-envtest.sh))
	mkdir -p $(ENVTEST_ASSETS_DIR)
	curl -sSLo $(ENVTEST_ASSETS_DIR)/setup-envtest.sh $(ENVTEST_SCRIPT_URL)
endif
	{ \
	source $(ENVTEST_ASSETS_DIR)/setup-envtest.sh && \
	fetch_envtest_tools $(ENVTEST_ASSETS_DIR) && \
	setup_envtest_env $(PWD)/$(ENVTEST_ASSETS_DIR) && \
	go test ./... -coverprofile cover.out ; \
	}

.PHONY: simple-test
simple-test:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr $$(go list ./...) 2>&1 | tee /dev/stderr)"
	go vet ./...

.PHONY: check-generate
check-generate:
	$(MAKE) manifests
	$(MAKE) generate
	git diff --exit-code --name-only

# Build manager binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -o bin/neco-admission -ldflags="-w -s" main.go

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=neco-admission webhook paths="./..."

# Generate code
.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Download controller-gen locally if necessary
CONTROLLER_GEN := $(PWD)/bin/controller-gen
.PHONY: controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_TOOLS_VERSION))

# Download kustomize locally if necessary
KUSTOMIZE := $(PWD)/bin/kustomize
.PHONY: kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v$(KUSTOMIZE_VERSION))

.PHONY: clean
clean:
	rm -rf bin testbin
	rm -f config/crd/bases/*

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
