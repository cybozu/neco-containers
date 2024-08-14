# Makefile for local-pv-provisioner

# Image version
IMAGE_VERSION = `cat ./TAG`
IMAGE_TAG = ghcr.io/cybozu/local-pv-provisioner:$(IMAGE_VERSION)
IMAGE_TAG_LATEST = ghcr.io/cybozu/local-pv-provisioner:latest

CONTROLLER_TOOLS_VERSION = 0.14.0
ENVTEST_K8S_VERSION = 1.29.3

# Specify envtest branch according to controller-runtime version
CONTROLLER_RUNTIME_VERSION := $(shell awk '$$1 == "sigs.k8s.io/controller-runtime" {print substr($$2, 2)}' go.mod)
ENVTEST_BRANCH := release-$(shell echo $(CONTROLLER_RUNTIME_VERSION) | cut -d "." -f 1-2)

# Set the shell used to bash for better error handling.
SHELL = /bin/bash
.SHELLFLAGS = -e -o pipefail -c

.PHONY: all
all: build

# Run tests, and set up envtest if not done already.
.PHONY: test
test: manifests generate simple-test setup-envtest
	{ \
	source <($(SETUP_ENVTEST) use -p env $(ENVTEST_K8S_VERSION)) && \
	go test -race -v -count 1 ./... ; \
	}

.PHONY: ci-e2e-test
ci-e2e-test:
	e2etest/ci-setup-kvm.sh
	$(MAKE) -C e2etest setup
	$(MAKE) -C e2etest test

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
build: manifests generate
	CGO_ENABLED=0 go build -o bin/local-pv-provisioner main.go

# Generate manifests e.g. RBAC etc.
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=local-pv-provisioner paths="./..."

# Generate code
.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Download controller-gen locally if necessary
CONTROLLER_GEN = $(CURDIR)/bin/controller-gen
.PHONY: controller-gen
controller-gen:
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_TOOLS_VERSION))

# Download setup-envtest locally if necessary
SETUP_ENVTEST = $(CURDIR)/bin/setup-envtest
.PHONY: setup-envtest
setup-envtest:
	$(call go-install-tool,$(SETUP_ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_BRANCH))

.PHONY: docker
docker: build
	docker build . -t $(IMAGE_TAG)
	docker tag $(IMAGE_TAG) $(IMAGE_TAG_LATEST)

.PHONY: clean
clean:
	rm -rf bin

include ../tool.mk
