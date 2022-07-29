# Makefile for local-pv-provisioner

IMAGE_VERSION = `cat ./TAG`
IMAGE_TAG = quay.io/cybozu/local-pv-provisioner:$(IMAGE_VERSION)
CONTROLLER_TOOLS_VERSION = 0.9.2
ENVTEST_K8S_VERSION = 1.24.2

# Set the shell used to bash for better error handling.
SHELL = /bin/bash
.SHELLFLAGS = -e -o pipefail -c

.PHONY: all
all: build

# Run tests, and set up envtest if not done already.
.PHONY: test
test: simple-test setup-envtest
	{ \
	source <($(SETUP_ENVTEST) use -p env $(ENVTEST_K8S_VERSION)) && \
	go test -race -v -count 1 ./... ; \
	}

.PHONY: simple-test
simple-test:
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
CONTROLLER_GEN := $(PWD)/bin/controller-gen
.PHONY: controller-gen
controller-gen:
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_TOOLS_VERSION))

# Download setup-envtest locally if necessary
SETUP_ENVTEST = $(PWD)/bin/setup-envtest
.PHONY: setup-envtest
setup-envtest:
	$(call go-install-tool,$(SETUP_ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

.PHONY: docker
docker: build
	docker build . -t $(IMAGE_TAG)

.PHONY: clean
clean:
	rm -rf bin

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
