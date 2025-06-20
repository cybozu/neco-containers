CILIUM_REPO ?= https://github.com/cybozu-go/cilium.git
CILIUM_TARGET ?= v$(shell cat TAG | cut -d'.' -f1-3)-cybozu
CILIUM_IMAGE_TOOLS_REPO ?= https://github.com/cilium/image-tools.git
CILIUM_IMAGE_TOOLS_TARGET ?= 3caa2c416a912dce00e641dd89201975be8c9ce8

IMAGE_TAG ?= ghcr.io/cybozu/cilium:$(shell cat TAG)

# ---- Development ----

.PHONY: checkout-repo
checkout-repo:
	@if [ -z "$(CHECKOUT_DIR)" ]; then \
		echo "Please specify CHECKOUT_DIR."; \
		exit 1; \
	fi
	@if [ -z "$(CHECKOUT_REPO)" ]; then \
		echo "Please specify CHECKOUT_REPO."; \
		exit 1; \
	fi
	@if [ -z "$(CHECKOUT_TARGET)" ]; then \
		echo "Please specify CHECKOUT_TARGET."; \
		exit 1; \
	fi
	if [ -d "$(CHECKOUT_DIR)" ]; then \
		echo "Skip: $(CHECKOUT_REPO) already checked out."; \
	else \
		mkdir -p $(CHECKOUT_DIR); \
		cd $(CHECKOUT_DIR); \
		git init 2>/dev/null; \
		git remote add origin $(CHECKOUT_REPO); \
		git fetch --depth 1 origin $(CHECKOUT_TARGET); \
		git checkout FETCH_HEAD 2>/dev/null; \
	fi

.PHONY: checkout-cilium
checkout-cilium:
	@CHECKOUT_DIR=src/cilium CHECKOUT_REPO=$(CILIUM_REPO) CHECKOUT_TARGET=$(CILIUM_TARGET) \
	$(MAKE) --no-print-directory checkout-repo

.PHONY: checkout-image-tools
checkout-image-tools:
	@CHECKOUT_DIR=src/image-tools CHECKOUT_REPO=$(CILIUM_IMAGE_TOOLS_REPO) CHECKOUT_TARGET=$(CILIUM_IMAGE_TOOLS_TARGET) \
	$(MAKE) --no-print-directory checkout-repo

.PHONY: checkout
checkout: checkout-cilium checkout-image-tools

# The following command displays the HEAD of the locally-checkoued-out upstream repositories.
.PHONY: status
status:
	@echo "cilium:      $$(cd src/cilium;      git log -n1 --decorate --pretty=oneline)"
	@echo "image-tools: $$(cd src/image-tools; git log -n1 --decorate --pretty=oneline)"

.PHONY: clean
clean:
	rm -rf src

.PHONY: build
build: checkout
	docker build . --tag=$(IMAGE_TAG)

# ---- Test & Investigation ----

# The following command shows tools version of the upstream image.
.PHONY: reference
reference:
	@IMAGE_TAG=quay.io/cilium/cilium:$(CILIUM_TARGET) \
	$(MAKE) --no-print-directory test

.PHONY: test
test:
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c 'cilium version'
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c 'bpftool version'
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c 'clang --version'
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c 'llc --version'
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c 'ip -V'
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c 'ss -V'
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c 'iptables -V'

.PHONY: test-e2e-setup
test-e2e-setup:
	@$(MAKE) --no-print-directory -C e2e setup

.PHONY: test-e2e
test-e2e-upgrade-inotify:
	@$(MAKE) --no-print-directory -C e2e upgrade-inotify

.PHONY: test-e2e
test-e2e:
	@$(MAKE) --no-print-directory -C e2e start
	@$(MAKE) --no-print-directory -C e2e test
	@$(MAKE) --no-print-directory -C e2e stop

# The following command inspects the files of the final container image.
# You can inspect any intermediate image by running `docker build . --target <AS_NAME> --tag <AS_NAME>:dev`.
.PHONY: dive
dive:
	@if ! which dive; then \
		echo '! Please install dive to inspect the container files.'; \
		echo '  https://github.com/wagoodman/dive'; \
		echo; \
		exit 1; \
	fi
	dive $(IMAGE_TAG)

# The following command displays the URLs of the upstream Dockerfiles to see.
.PHONY: urls
urls:
	@echo "https://github.com/cilium/image-tools/blob/$(CILIUM_IMAGE_TOOLS_TARGET)/images/compilers/Dockerfile"
	@echo "https://github.com/cilium/image-tools/blob/$(CILIUM_IMAGE_TOOLS_TARGET)/images/llvm/Dockerfile"
	@echo "https://github.com/cilium/image-tools/blob/$(CILIUM_IMAGE_TOOLS_TARGET)/images/bpftool/Dockerfile"
	@echo "https://github.com/cilium/image-tools/blob/$(CILIUM_IMAGE_TOOLS_TARGET)/images/iptables/Dockerfile"
	@echo "https://github.com/cilium/cilium/blob/$(CILIUM_TARGET)/images/runtime/Dockerfile"
	@echo "https://github.com/cilium/cilium/blob/$(CILIUM_TARGET)/images/builder/Dockerfile"
	@echo "https://github.com/cilium/cilium/blob/$(CILIUM_TARGET)/images/cilium/Dockerfile"
