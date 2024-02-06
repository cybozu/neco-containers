SUDO = sudo

.PHONY: all
all: check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: test
test: staticcheck custom-checker
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	$(STATICCHECK) ./...
	test -z "$$($(CUSTOM_CHECKER) -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	go test -c ./...
	$(SUDO) ./bpf-map-pressure-exporter.test -test.v
	rm -f ./bpf-map-pressure-exporter.test

include ../tool.mk
