CONTROLLER_TOOLS_VERSION = 0.6.2
KUSTOMIZE_VERSION = 4.2.0
CALICO_VERSION = 3.20
CONTOUR_VERSION = 1.18.1
ARGOCD_VERSION = 2.1.2
GRAFANA_OPERATOR_VERSION = 3.10.4
ENVTEST_K8S_VERSION = 1.21.4

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
	curl -fsL -o config/crd/bases/grafanadashboard.yaml https://github.com/grafana-operator/grafana-operator/raw/v$(GRAFANA_OPERATOR_VERSION)/deploy/crds/GrafanaDashboard.yaml

# Run tests, and set up envtest if not done already.
.PHONY: test
test: crds simple-test setup-envtest
	{ \
	source <($(SETUP_ENVTEST) use -p env $(ENVTEST_K8S_VERSION)) && \
	go test ./... -coverprofile cover.out ; \
	}
	# TEST_PERMISSIVE test and NO_HTTPPROXY_MUTATION test can run at the same time
	# because they do not affect each other.
	{ \
	source <($(SETUP_ENVTEST) use -p env $(ENVTEST_K8S_VERSION)) && \
	TEST_PERMISSIVE=true NO_HTTPPROXY_MUTATION=true go test -v -count 1 ./... ; \
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
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_TOOLS_VERSION))

# Download kustomize locally if necessary
KUSTOMIZE := $(PWD)/bin/kustomize
.PHONY: kustomize
kustomize:
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v$(KUSTOMIZE_VERSION))

# Download setup-envtest locally if necessary
SETUP_ENVTEST = $(PWD)/bin/setup-envtest
.PHONY: setup-envtest
setup-envtest:
	$(call go-install-tool,$(SETUP_ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

.PHONY: clean
clean:
	rm -rf bin
	rm -f config/crd/bases/*

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
