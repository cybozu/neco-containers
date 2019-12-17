# Makefile for local-pv-provisioner

IMAGE_VERSION = `cat ./TAG`
IMAGE_TAG = quay.io/cybozu/local-pv-provisioner:$(IMAGE_VERSION)
KUBEBUILDER_VERSION = 2.2.0
CTRLTOOLS_VERSION = 0.2.4
KIND_VERSION = 0.6.1
KIND_NODE_VERSION = 1.16
KIND_NAME = local-pv-provisioner-test
KUBERNETES_VERSION= 1.16.1

SUDO=sudo
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
	test -z "$$(restrictpkg -packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	ineffassign .
	go test -race -v ./...
	go vet ./...

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

launch-kind:
	sed s/@KUBERNETES_VERSION@/$(KUBERNETES_VERSION)/ cluster.yaml > /tmp/local-pv-provisioner/cluster.yaml
	env KUBECONFIG= kind create cluster --config /tmp/local-pv-provisioner/cluster.yaml --image quay.io/cybozu/kind-node:$(KIND_NODE_VERSION) --name=$(KIND_NAME) --wait=300s

load-image-to-kind: launch-kind
	kind load docker-image $(IMAGE_TAG) -q --name=$(KIND_NAME)

shutdown-kind:
	kind delete cluster --name=local-pv-provisioner-test || true

setup:
	curl -sfL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/
	$(SUDO) rm -rf /usr/local/kubebuilder
	$(SUDO) mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) /usr/local/kubebuilder
	$(SUDO) curl -o /usr/local/kubebuilder/bin/kustomize -sL https://go.kubebuilder.io/kustomize/$(GOOS)/$(GOARCH)
	$(SUDO) chmod a+x /usr/local/kubebuilder/bin/kustomize
	cd /tmp; GO111MODULE=on GOFLAGS= go get sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CTRLTOOLS_VERSION)
	cd /tmp; GO111MODULE=on go get sigs.k8s.io/kind@v$(KIND_VERSION)

clean:
	rm -rf bin

.PHONY: all test build manifests generate clean docker launch-kind load-image-to-kind shutdown-kind setup
