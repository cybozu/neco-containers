.PHONY: all
all: check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	cd tools && go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint:
	go tool -modfile=tools/go.mod golangci-lint run ./... -vvv

.PHONY: lint-fix
lint-fix:
	go tool -modfile=tools/go.mod golangci-lint run ./... -vvv --fix

.PHONY: test
test: clean
	$(MAKE) lint
	go tool -modfile=tools/go.mod ginkgo -v --race -p .

.PHONY: clean
clean:
	rm -f testdata/output_main_test/*
	rm -fr testdata/pointers_get_machines
	rm -fr testdata/pointers_log_collector
	rm -fr testdata/pointers_main_test
