.PHONY: all
all: check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: test
test: clean
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	go tool staticcheck ./...
	test -z "$$(go tool custom-checker -restrictpkg.packages=html/template ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	go tool ginkgo -v --race -p .

.PHONY: clean
clean:
	rm -f testdata/output_main_test/*
	rm -fr testdata/pointers_get_machines
	rm -fr testdata/pointers_log_collector
	rm -fr testdata/pointers_main_test
