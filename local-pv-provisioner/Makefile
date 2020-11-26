# Makefile for local-pv-provisioner

IMAGE_VERSION = `cat ./TAG`
IMAGE_TAG = quay.io/cybozu/local-pv-provisioner:$(IMAGE_VERSION)
KUBEBUILDER_VERSION = 2.3.1
CTRLTOOLS_VERSION = 0.4.0

SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOFLAGS = -mod=vendor
export GOFLAGS

.PHONY: all
all: build

# Run tests
.PHONY: test
test:
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | grep -v vendor/ | tee /dev/stderr)"
	ineffassign .
	go vet ./...
	$(SUDO_GO) test -race -v ./...

# Build manager binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -o bin/local-pv-provisioner main.go

# Generate manifests e.g. RBAC etc.
.PHONY: manifests
manifests:
	controller-gen rbac:roleName=local-pv-provisioner paths="./..."

# Generate code
.PHONY: generate
generate:
	controller-gen object:headerFile=./hack/boilerplate.go.txt paths="./..."

.PHONY: check-uncommitted
check-uncommitted:
	$(MAKE) manifests
	$(MAKE) generate
	go mod tidy
	git diff --exit-code --name-only

.PHONY: docker
docker: build
	docker build . -t $(IMAGE_TAG)

.PHONY: setup
setup:
	curl -sfL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/
	$(SUDO) rm -rf /usr/local/kubebuilder
	$(SUDO) mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) /usr/local/kubebuilder
	cd /tmp; GO111MODULE=on GOFLAGS= go get sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CTRLTOOLS_VERSION)

.PHONY: clean
clean:
	rm -rf bin
