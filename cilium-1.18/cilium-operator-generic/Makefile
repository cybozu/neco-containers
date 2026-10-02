# The fork repository and the commit SHA are kept in Dockerfile so that
# `docker build` alone is reproducible. Read them from there to avoid
# having two sources of truth.
CILIUM_REPO := $(shell sed -n 's/^ENV CILIUM_REPOSITORY=//p' Dockerfile)
CILIUM_TARGET := $(shell sed -n 's/^ENV CILIUM_SHA=//p' Dockerfile)
CILIUM_FORK_BRANCH := v$(shell cat TAG | cut -d'.' -f1-3)-cybozu

.PHONY: check-cilium-target
check-cilium-target:
	@if [ -z "$(CILIUM_REPO)" ] || [ -z "$(CILIUM_TARGET)" ]; then \
		echo "Failed to read ENV CILIUM_REPOSITORY / CILIUM_SHA from Dockerfile."; \
		exit 1; \
	fi; \
	LATEST=$$(git ls-remote $(CILIUM_REPO) refs/heads/$(CILIUM_FORK_BRANCH) | cut -f1); \
	if [ -z "$$LATEST" ]; then \
		echo "Failed to resolve branch $(CILIUM_FORK_BRANCH) on $(CILIUM_REPO)."; \
		exit 1; \
	fi; \
	if [ "$$LATEST" != "$(CILIUM_TARGET)" ]; then \
		echo "CILIUM_SHA ($(CILIUM_TARGET)) is not the tip of $(CILIUM_FORK_BRANCH) ($$LATEST)."; \
		echo "Update ENV CILIUM_SHA in $(CURDIR)/Dockerfile to $$LATEST."; \
		exit 1; \
	fi
