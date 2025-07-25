CONTROLLER_TOOLS_VERSION = 0.18.0
KUSTOMIZE_VERSION = 5.7.0
CONTOUR_VERSION = 1.30.1
ARGOCD_VERSION = 2.11.7
GRAFANA_OPERATOR_VERSION = 5.10.0
ENVTEST_K8S_VERSION = 1.32.0

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
	curl -fsL -o config/crd/bases/grafana-operator.yaml https://github.com/grafana/grafana-operator/raw/v$(GRAFANA_OPERATOR_VERSION)/deploy/kustomize/base/crds.yaml

# Run tests, and set up envtest if not done already.
.PHONY: test
test: crds simple-test setup-envtest
	{ \
	source <($(SETUP_ENVTEST) use -p env $(ENVTEST_K8S_VERSION)) && \
	ENABLE_PODCPUREQUESTREDUCE_MUTATION=true go test ./... -coverprofile cover.out ; \
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
CONTROLLER_GEN = $(CURDIR)/bin/controller-gen
.PHONY: controller-gen
controller-gen:
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_TOOLS_VERSION))

# Download kustomize locally if necessary
KUSTOMIZE = $(CURDIR)/bin/kustomize
.PHONY: kustomize
kustomize:
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v$(shell echo $(KUSTOMIZE_VERSION) | cut -d "." -f1)@v$(KUSTOMIZE_VERSION))

# Download setup-envtest locally if necessary
SETUP_ENVTEST = $(CURDIR)/bin/setup-envtest
.PHONY: setup-envtest
setup-envtest:
	$(call go-install-tool,$(SETUP_ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_BRANCH))

.PHONY: clean
clean:
	rm -rf bin
	rm -f config/crd/bases/*

include ../tool.mk
