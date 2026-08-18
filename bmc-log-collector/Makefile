AQUA_POLICY_CONFIG := $(CURDIR)/aqua-policy.yaml
export AQUA_POLICY_CONFIG

.PHONY: all
all: check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint: setup
	golangci-lint run ./... -vvv

.PHONY: lint-fix
lint-fix: setup
	golangci-lint run ./... -vvv --fix

.PHONY: test
test: clean
	$(MAKE) lint
	ginkgo -v --race -p .

.PHONY: setup
setup:
	aqua install -l

.PHONY: clean
clean:
	rm -f testdata/output_main_test/*
	rm -fr testdata/pointers_get_machines
	rm -fr testdata/pointers_log_collector
	rm -fr testdata/pointers_main_test
