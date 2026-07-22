BCC_REPO ?= https://github.com/iovisor/bcc.git
BCC_VERSION ?= 0.37.0  # used as -DREVISION in cmake to build bcc
BCC_SHA ?= 306a2819f73d9525430693e6399d58caf6e12b3b  # v0.37.0
LIBBPF_REPO ?= https://github.com/libbpf/libbpf.git
LIBBPF_SHA ?= f5dcbae736e5d7f83a35718e01be1a8e3010fa39  # v1.7.0
BPFTRACE_REPO ?= https://github.com/bpftrace/bpftrace.git
BPFTRACE_SHA ?= 1bac8a8daad6e0aa579980e877d0c9d5fc296593  # v0.26.1

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
	@if [ -z "$(CHECKOUT_SHA)" ]; then \
		echo "Please specify CHECKOUT_SHA."; \
		exit 1; \
	fi
	if [ -d "$(CHECKOUT_DIR)" ]; then \
		echo "Skip: $(CHECKOUT_REPO) already checked out."; \
	else \
		mkdir -p $(CHECKOUT_DIR); \
		cd $(CHECKOUT_DIR); \
		git init 2>/dev/null; \
		git remote add origin $(CHECKOUT_REPO); \
		git fetch --depth 1 origin $(CHECKOUT_SHA); \
		git checkout --detach "$(CHECKOUT_SHA)"; \
		git log -n1 --oneline; \
	fi

.PHONY: checkout
checkout:
	@CHECKOUT_DIR=src/bcc CHECKOUT_REPO=$(BCC_REPO) CHECKOUT_SHA=$(BCC_SHA) \
	$(MAKE) --no-print-directory checkout-repo

	@CHECKOUT_DIR=src/libbpf CHECKOUT_REPO=$(LIBBPF_REPO) CHECKOUT_SHA=$(LIBBPF_SHA) \
	$(MAKE) --no-print-directory checkout-repo

	@CHECKOUT_DIR=src/bpftrace CHECKOUT_REPO=$(BPFTRACE_REPO) CHECKOUT_SHA=$(BPFTRACE_SHA) \
	$(MAKE) --no-print-directory checkout-repo

.PHONY: clean
clean:
	rm -rf src

.PHONY: build
build:
	docker build . --progress=plain --target builder -t $(BUILDER_IMAGE_TAG)
	docker build . --progress=plain -t $(IMAGE_TAG)

.PHONY: bcc-version
bcc-version:
	@echo $(BCC_VERSION)

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
