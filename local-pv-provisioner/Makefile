# Makefile for local-pv-provisioner

IMAGE_VERSION = `cat ./TAG`
IMAGE_TAG = quay.io/cybozu/local-pv-provisioner:$(IMAGE_VERSION)
KUBEBUILDER_VERSION = 2.2.0
CTRLTOOLS_VERSION = 0.2.4

SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOFLAGS = -mod=vendor
export GOFLAGS

all: build

# Run tests
test:
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
	test -z "$$(golint $$(go list ./... | grep -v /vendor/) | grep -v '/mtest/.*: should not use dot imports' | tee /dev/stderr)"
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | grep -v /vendor/ | tee /dev/stderr)"
	ineffassign .
	go vet ./...
	$(SUDO_GO) test -race -v ./...

# Build manager binary
build:
	CGO_ENABLED=0 go build -o bin/local-pv-provisioner main.go

# Generate manifests e.g. RBAC etc.
manifests:
	controller-gen rbac:roleName=local-pv-provisioner paths="./..."

# Generate code
generate:
	controller-gen object:headerFile=./hack/boilerplate.go.txt paths="./..."

docker: build
	docker build . -t $(IMAGE_TAG)

setup:
	curl -sfL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/
	$(SUDO) rm -rf /usr/local/kubebuilder
	$(SUDO) mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) /usr/local/kubebuilder
	cd /tmp; GO111MODULE=on GOFLAGS= go get sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CTRLTOOLS_VERSION)

clean:
	rm -rf bin

.PHONY: all test build manifests generate docker setup clean
