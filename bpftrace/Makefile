BCC_REPO ?= https://github.com/iovisor/bcc.git
BCC_TARGET ?= v0.33.0
LIBBPF_REPO ?= https://github.com/libbpf/libbpf.git
LIBBPF_TARGET ?= v1.5.0
BPFTRACE_REPO ?= https://github.com/bpftrace/bpftrace.git
BPFTRACE_TARGET ?= v$(shell cat TAG | cut -d. -f1-3)

BUILDER_IMAGE_TAG ?= bpftrace-builder:dev
IMAGE_TAG ?= ghcr.io/cybozu/bpftrace:$(shell cat TAG)

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
	@CHECKOUT_DIR=src/bcc CHECKOUT_REPO=$(BCC_REPO) CHECKOUT_TARGET=$(BCC_TARGET) \
	$(MAKE) --no-print-directory checkout-repo

	@CHECKOUT_DIR=src/libbpf CHECKOUT_REPO=$(LIBBPF_REPO) CHECKOUT_TARGET=$(LIBBPF_TARGET) \
	$(MAKE) --no-print-directory checkout-repo

	@CHECKOUT_DIR=src/bpftrace CHECKOUT_REPO=$(BPFTRACE_REPO) CHECKOUT_TARGET=$(BPFTRACE_TARGET) \
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
	docker run --rm $(IMAGE_TAG) bpftrace --version

.PHONY: try
try:
	docker run --rm --privileged --pid=host --net=host \
		-v /lib/modules/:/lib/modules:ro \
		-v /sys/kernel/debug/:/sys/kernel/debug:rw \
		$(IMAGE_TAG) bpftrace -l | grep do_nanosleep
	@echo
	docker run --rm --privileged --pid=host --net=host \
		-v /lib/modules/:/lib/modules:ro \
		-v /sys/kernel/debug/:/sys/kernel/debug:rw \
		$(IMAGE_TAG) bpftrace -e 'kfunc:vmlinux:do_nanosleep { printf("%d, %s\n", pid, comm); if (++@x >= 10) { exit(); } }'
