CONTROLLER_TOOLS_VERSION = 0.15.0
KUSTOMIZE_VERSION = 5.4.1
CONTOUR_VERSION = 1.28.2
ARGOCD_VERSION = 2.10.7
GRAFANA_OPERATOR_VERSION = 5.8.1
ENVTEST_K8S_VERSION = 1.28.3

# Specify envtest branch according to controller-runtime version
CONTROLLER_RUNTIME_VERSION := $(shell awk '$$1 == "sigs.k8s.io/controller-runtime" {print substr($$2, 2)}' go.mod)
ENVTEST_BRANCH := release-$(shell echo $(CONTROLLER_RUNTIME_VERSION) | cut -d "." -f 1-2)

# Set the shell used to bash for better error handling.
SHELL = /bin/bash
.SHELLFLAGS = -e -o pipefail -c

.PHONY: all
all: build

.PHONY: crds
crds:
	mkdir -p config/crd/bases
	curl -fsL -o config/crd/bases/contour.yaml https://raw.githubusercontent.com/projectcontour/contour/v$(CONTOUR_VERSION)/examples/contour/01-crds.yaml
	curl -fsL -o config/crd/bases/application.yaml https://raw.githubusercontent.com/argoproj/argo-cd/v$(ARGOCD_VERSION)/manifests/crds/application-crd.yaml
	curl -fsL -o config/crd/bases/grafana-operator.yaml https://github.com/grafana-operator/grafana-operator/raw/v$(GRAFANA_OPERATOR_VERSION)/deploy/kustomize/base/crds.yaml

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
simple-test: staticcheck
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	$(STATICCHECK) ./...
	go vet ./...

.PHONY: check-generate
check-generate:
	$(MAKE) manifests
	$(MAKE) generate
	go mod tidy
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
CONTROLLER_GEN = $(PROJECT_DIR)/bin/controller-gen
.PHONY: controller-gen
controller-gen:
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_TOOLS_VERSION))

# Download kustomize locally if necessary
KUSTOMIZE = $(PROJECT_DIR)/bin/kustomize
.PHONY: kustomize
kustomize:
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v$(shell echo $(KUSTOMIZE_VERSION) | cut -d "." -f1)@v$(KUSTOMIZE_VERSION))

# Download setup-envtest locally if necessary
SETUP_ENVTEST = $(PROJECT_DIR)/bin/setup-envtest
.PHONY: setup-envtest
setup-envtest:
	$(call go-install-tool,$(SETUP_ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_BRANCH))

# Download staticcheck locally if necessary
STATICCHECK = $(PROJECT_DIR)/bin/staticcheck
.PHONY: staticcheck
staticcheck:
	$(call go-install-tool,$(STATICCHECK),honnef.co/go/tools/cmd/staticcheck@latest)

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
