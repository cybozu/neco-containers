# Contour v1.33.5
CONTOUR_COMMIT_HASH := 2713ad0e0ebd5d192a53803cef4c1c3a69692b1c
# Argo CD v3.4.3
ARGOCD_COMMIT_HASH := 1801122b4391cad4961301f787006dc9a88c2dd3
# Grafana Operator v5.18.0
GRAFANA_OPERATOR_COMMIT_HASH := de1156accb9c4529cbdb536a054b9c4bb5ee759d

ENVTEST_K8S_VERSION := 1.35.0
ENVTEST_USE := setup-envtest use -p env $(ENVTEST_K8S_VERSION)

# Set the shell used to bash for better error handling.
SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

.PHONY: all
all: build

.PHONY: crds
crds:
	mkdir -p config/crd/bases
	curl -fsL -o config/crd/bases/contour.yaml https://raw.githubusercontent.com/projectcontour/contour/$(CONTOUR_COMMIT_HASH)/examples/contour/01-crds.yaml
	curl -fsL -o config/crd/bases/application.yaml https://raw.githubusercontent.com/argoproj/argo-cd/$(ARGOCD_COMMIT_HASH)/manifests/crds/application-crd.yaml
	curl -fsL -o config/crd/bases/grafana-operator.yaml https://raw.githubusercontent.com/grafana/grafana-operator/$(GRAFANA_OPERATOR_COMMIT_HASH)/deploy/kustomize/base/crds.yaml

# Run tests, and set up envtest if not done already.
.PHONY: test
test: crds simple-test setup
	source <($(ENVTEST_USE)) && \
		ENABLE_PODCPUREQUESTREDUCE_MUTATION=true go test ./... -coverprofile cover.out
	# TEST_PERMISSIVE test and NO_HTTPPROXY_MUTATION test can run at the same time
	# because they do not affect each other.
	source <($(ENVTEST_USE)) && \
		TEST_PERMISSIVE=true NO_HTTPPROXY_MUTATION=true go test -v -count 1 ./...

.PHONY: simple-test
simple-test: setup
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
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
manifests: setup
	controller-gen rbac:roleName=neco-admission webhook paths="./..."

# Generate code
.PHONY: generate
generate: setup
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: clean
clean:
	rm -rf bin
	rm -f config/crd/bases/*

.PHONY: setup
setup:
	aqua i -l
