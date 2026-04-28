HUBBLE_UI_REPO ?= https://github.com/cilium/hubble-ui.git
HUBBLE_UI_TARGET ?= v$(shell cat TAG | cut -d'.' -f1-3)
NGINX_REPO ?= https://github.com/nginx/docker-nginx-unprivileged.git
NGINX_COMMIT_HASH ?= 872c2c0617f10eecc31b7826168a759f24dc02d7
PNPM_VERSION ?= 10.33.2
PNPM_SHA ?= 39d7b6600239712bc9581ea219b17ffef46ba60998779cb717be2e068be029ef
PNPM_BIN ?= $(CURDIR)/.pnpm-bin/pnpm

BACKEND_BUILDER_IMAGE_TAG ?= backend-builder:dev
BACKEND_IMAGE_TAG ?= ghcr.io/cybozu/hubble-ui-backend:$(shell cat TAG)
FRONTEND_BUILDER_IMAGE_TAG ?= frontend-builder:dev
FRONTEND_IMAGE_TAG ?= ghcr.io/cybozu/hubble-ui-frontend:$(shell cat TAG)

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

.PHONY: checkout
checkout:
	@CHECKOUT_DIR=src/hubble-ui CHECKOUT_REPO=$(HUBBLE_UI_REPO) CHECKOUT_TARGET=$(HUBBLE_UI_TARGET) \
	$(MAKE) --no-print-directory checkout-repo

	@CHECKOUT_DIR=src/docker-nginx-unprivileged CHECKOUT_REPO=$(NGINX_REPO) CHECKOUT_TARGET=$(NGINX_COMMIT_HASH) \
	$(MAKE) --no-print-directory checkout-repo

.PHONY: clean
clean:
	rm -rf src

.PHONY: build-backend
build-backend:
	docker build . --progress=plain --target backend-builder -t $(BACKEND_BUILDER_IMAGE_TAG)
	docker build . --progress=plain --target backend -t $(BACKEND_IMAGE_TAG)

.PHONY: build-frontend
build-frontend: pnpm-bin
	docker build . --progress=plain --target frontend-builder -t $(FRONTEND_BUILDER_IMAGE_TAG)
	docker build . --progress=plain --target frontend -t $(FRONTEND_IMAGE_TAG)

.PHONY: pnpm-bin
pnpm-bin:
	@mkdir -p $(dir $(PNPM_BIN))
	@echo "$(PNPM_SHA)  $(PNPM_BIN)" | sha256sum -c --status 2>/dev/null || ( \
	   curl -fsSL -o $(PNPM_BIN) https://github.com/pnpm/pnpm/releases/download/v$(PNPM_VERSION)/pnpm-linux-x64 \
	   && echo "$(PNPM_SHA)  $(PNPM_BIN)" | sha256sum -c - \
	   && chmod +x $(PNPM_BIN) )

.PHONY: import-lockfile
import-lockfile: checkout pnpm-bin
	cd src/hubble-ui && $(PNPM_BIN) import
	cp src/hubble-ui/pnpm-lock.yaml ./pnpm-lock.yaml

.PHONY: verify-lockfile
verify-lockfile: checkout pnpm-bin
	cd src/hubble-ui && $(PNPM_BIN) import >/dev/null
	@if ! diff -q ./pnpm-lock.yaml src/hubble-ui/pnpm-lock.yaml >/dev/null; then \
	   echo "pnpm-lock.yaml is out of date. Run 'make import-lockfile' and commit the result."; \
	   exit 1; \
	 fi
