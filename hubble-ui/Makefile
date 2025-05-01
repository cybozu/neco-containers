HUBBLE_UI_REPO ?= https://github.com/cilium/hubble-ui.git
HUBBLE_UI_TARGET ?= v$(shell cat TAG | cut -d'.' -f1-3)
NGINX_REPO ?= https://github.com/nginx/docker-nginx-unprivileged.git
NGINX_COMMIT_HASH ?= 7e515bf70e10e5ea7cdcc14fd3e736ebc639df9a

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
build-frontend:
	docker build . --progress=plain --target frontend-builder -t $(FRONTEND_BUILDER_IMAGE_TAG)
	docker build . --progress=plain --target frontend -t $(FRONTEND_IMAGE_TAG)
