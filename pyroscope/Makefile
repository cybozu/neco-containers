PYROSCOPE_REPO ?= https://github.com/grafana/pyroscope.git
PYROSCOPE_TARGET ?= v$(shell cat TAG | cut -d'.' -f1-3)

BUILDER_IMAGE_TAG ?= pyroscope-builder:dev
IMAGE_TAG ?= ghcr.io/cybozu/pyroscope:$(shell cat TAG)

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
	@CHECKOUT_DIR=src/pyroscope CHECKOUT_REPO=$(PYROSCOPE_REPO) CHECKOUT_TARGET=$(PYROSCOPE_TARGET) \
	$(MAKE) --no-print-directory checkout-repo

.PHONY: clean
clean:
	rm -rf src

.PHONY: build
build:
	docker build . --progress=plain --target builder -t $(BUILDER_IMAGE_TAG)
	docker build . --progress=plain -t $(IMAGE_TAG)

.PHONY: test
test:
	docker run --rm --entrypoint pyroscope $(IMAGE_TAG) -h
