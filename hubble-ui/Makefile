HUBBLE_UI_REPO ?= https://github.com/cilium/hubble-ui.git
HUBBLE_UI_COMMIT_HASH ?= d6eb0dd9ff8e684a6a6bd6da7b6cb304497a2e26
NGINX_REPO ?= https://github.com/nginx/docker-nginx-unprivileged.git
NGINX_COMMIT_HASH ?= ecb807035c7963a87b2a88d339460b1e8eeaec2f
PNPM_DIR ?= $(CURDIR)/.pnpm-bin
PNPM_BIN ?= $(CURDIR)/.pnpm-bin/pnpm
TAKUMI_GUARD_REGISTRY ?= https://npm.flatt.tech/
# Do not resolve versions published less than 3 days (4320 minutes) ago.
# Matches the company-wide npm / pnpm security guideline.
PNPM_MINIMUM_RELEASE_AGE ?= 4320

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
	@CHECKOUT_DIR=src/hubble-ui CHECKOUT_REPO=$(HUBBLE_UI_REPO) CHECKOUT_TARGET=$(HUBBLE_UI_COMMIT_HASH) \
	$(MAKE) --no-print-directory checkout-repo

	@CHECKOUT_DIR=src/docker-nginx-unprivileged CHECKOUT_REPO=$(NGINX_REPO) CHECKOUT_TARGET=$(NGINX_COMMIT_HASH) \
	$(MAKE) --no-print-directory checkout-repo

.PHONY: clean
clean:
	rm -rf src $(PNPM_DIR)

.PHONY: build-backend
build-backend:
	docker build . --progress=plain --target backend-builder -t $(BACKEND_BUILDER_IMAGE_TAG)
	docker build . --progress=plain --target backend -t $(BACKEND_IMAGE_TAG)

.PHONY: build-frontend
build-frontend: $(PNPM_BIN)
	docker build . --progress=plain --target frontend-builder -t $(FRONTEND_BUILDER_IMAGE_TAG)
	docker build . --progress=plain --target frontend -t $(FRONTEND_IMAGE_TAG)

.PHONY: pnpm-bin
pnpm-bin: $(PNPM_BIN)

$(PNPM_BIN):
	aqua install
	mkdir -p $(PNPM_DIR)
	cp -r $$(dirname $$(aqua which pnpm))/* $(PNPM_DIR)

# Refuse to proceed unless pnpm is configured to fetch through Takumi Guard.
.PHONY: check-takumi-guard
check-takumi-guard: $(PNPM_BIN)
	@if [ "$$($(PNPM_BIN) config get registry)" != "$(TAKUMI_GUARD_REGISTRY)" ]; then \
	    echo "ERROR: pnpm registry is not Takumi Guard ($(TAKUMI_GUARD_REGISTRY)). Setup: https://shisho.dev/docs/ja/t/guard/quickstart/npm/" >&2; \
		exit 1; \
	fi

# .pnpmfile.cjs pins packages to src/hubble-ui/package-lock.json (see its
# header), so the lockfile follows HUBBLE_UI_COMMIT_HASH and Dockerfile can
# run with --frozen-lockfile.
#
# `clean` is a hard dep so that every run resolves from a fresh checkout.
#
# A full install (not --lockfile-only) routes every tarball through Takumi
# Guard so the new dep set is screened before the lockfile is committed.
#
# minimumReleaseAge is passed explicitly instead of relying on the pnpm 11
# default (1 day) so the waiting period stays 3 days even if the default
# changes. Upstream ships its own .npmrc, so we do not copy one in.
.PHONY: import-lockfile
import-lockfile: clean check-takumi-guard checkout
	cp ./.pnpmfile.cjs src/hubble-ui/.pnpmfile.cjs
	cd src/hubble-ui && $(PNPM_BIN) install --no-frozen-lockfile --ignore-scripts \
		--config.minimumReleaseAge=$(PNPM_MINIMUM_RELEASE_AGE)
	cp src/hubble-ui/pnpm-lock.yaml ./pnpm-lock.yaml

.PHONY: verify-lockfile
verify-lockfile: clean $(PNPM_BIN) checkout
# https://google.github.io/osv-scanner/output/#return-codes
	@result=$$(mktemp); \
	osv-scanner scan source --lockfile=pnpm-lock.yaml --format=markdown > "$${result}"; \
	rc=$$?; \
	if [ "$${rc}" -ne 0 ] && [ "$${rc}" -ne 1 ]; then \
		echo "unexpected error"; \
		cat "$${result}"; \
		exit "$${rc}"; \
	fi; \
	if grep -q MAL- "$${result}"; then \
		echo '!! Possible malicious dependency. Please investigate the root cause.'; \
		cat "$${result}"; \
		exit 1; \
	fi; \
	rm "$${result}"

	cp ./pnpm-lock.yaml ./.pnpmfile.cjs src/hubble-ui/
	cd src/hubble-ui && $(PNPM_BIN) install --frozen-lockfile --lockfile-only --ignore-scripts
