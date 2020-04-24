# Makefile for custom Kubernetes admission controllers

KUBEBUILDER_VERSION = 2.3.0
CTRLTOOLS_VERSION = 0.2.8
CALICO_VERSION = 3.11
CONTOUR_VERSION = 1.3.0
ARGOCD_VERSION = 1.5.2
GRAFANA_OPERATOR_VERSION = 3.2.0

SUDO=sudo
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOFLAGS = -mod=vendor
export GOFLAGS

all: build

# Run tests
test: crds
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
	test -z "$$(golint $$(go list ./... | grep -v /vendor/) | grep -v '/mtest/.*: should not use dot imports' | tee /dev/stderr)"
	test -z "$$(nilerr $$(go list ./... | grep -v /vendor/) 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | grep -v vendor/ | tee /dev/stderr)"
	ineffassign .
	go test -race -v ./...
	go vet ./...

# Build manager binary
build:
	CGO_ENABLED=0 go build -o bin/neco-admission main.go

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	controller-gen rbac:roleName=neco-admission webhook paths="./..."

# Generate code
generate:
	controller-gen object:headerFile=./hack/boilerplate.go.txt paths="./..."

setup:
	curl -sfL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/
	$(SUDO) rm -rf /usr/local/kubebuilder
	$(SUDO) mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) /usr/local/kubebuilder
	$(SUDO) curl -o /usr/local/kubebuilder/bin/kustomize -sL https://go.kubebuilder.io/kustomize/$(GOOS)/$(GOARCH)
	$(SUDO) chmod a+x /usr/local/kubebuilder/bin/kustomize
	cd /tmp; GO111MODULE=on GOFLAGS= go get sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CTRLTOOLS_VERSION)

crds:
	mkdir -p config/crd/bases
	curl -fsL -o config/crd/bases/calico.yaml https://docs.projectcalico.org/v$(CALICO_VERSION)/manifests/crds.yaml
	curl -fsL -o config/crd/bases/contour.yaml https://raw.githubusercontent.com/projectcontour/contour/v$(CONTOUR_VERSION)/examples/contour/01-crds.yaml
	curl -fsL -o config/crd/bases/application.yaml https://raw.githubusercontent.com/argoproj/argo-cd/v$(ARGOCD_VERSION)/manifests/crds/application-crd.yaml
	curl -fsL -o config/crd/bases/grafanadashboard.yaml https://github.com/integr8ly/grafana-operator/raw/v$(GRAFANA_OPERATOR_VERSION)/deploy/crds/GrafanaDashboard.yaml

clean:
	rm -rf bin

.PHONY: all test build manifests generate clean setup crds
