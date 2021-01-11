# Makefile for custom Kubernetes admission controllers

KUBEBUILDER_VERSION = 2.3.0
CTRLTOOLS_VERSION = 0.2.8
CALICO_VERSION = 3.16
CONTOUR_VERSION = 1.9.0
ARGOCD_VERSION = 1.7.7
GRAFANA_OPERATOR_VERSION = 3.7.0

SUDO=sudo
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

.PHONY: all
all: build

# Run tests
.PHONY: test
test: crds
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr $$(go list ./...) 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	go test -race -v ./...


# Build manager binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -o bin/neco-admission main.go

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests:
	controller-gen rbac:roleName=neco-admission webhook paths="./..."

# Generate code
.PHONY: generate
generate:
	controller-gen object:headerFile=./hack/boilerplate.go.txt paths="./..."

.PHONY: setup
setup:
	curl -sfL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/
	$(SUDO) rm -rf /usr/local/kubebuilder
	$(SUDO) mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) /usr/local/kubebuilder
	$(SUDO) curl -o /usr/local/kubebuilder/bin/kustomize -sL https://go.kubebuilder.io/kustomize/$(GOOS)/$(GOARCH)
	$(SUDO) chmod a+x /usr/local/kubebuilder/bin/kustomize
	cd /tmp; GO111MODULE=on GOFLAGS= go get sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CTRLTOOLS_VERSION)

.PHONY: crds
crds:
	mkdir -p config/crd/bases
	curl -fsL -o config/crd/bases/calico.yaml https://docs.projectcalico.org/v$(CALICO_VERSION)/manifests/crds.yaml
	curl -fsL -o config/crd/bases/contour.yaml https://raw.githubusercontent.com/projectcontour/contour/v$(CONTOUR_VERSION)/examples/contour/01-crds.yaml
	curl -fsL -o config/crd/bases/application.yaml https://raw.githubusercontent.com/argoproj/argo-cd/v$(ARGOCD_VERSION)/manifests/crds/application-crd.yaml
	curl -fsL -o config/crd/bases/grafanadashboard.yaml https://github.com/integr8ly/grafana-operator/raw/v$(GRAFANA_OPERATOR_VERSION)/deploy/crds/GrafanaDashboard.yaml

.PHONY: clean
clean:
	rm -rf bin
