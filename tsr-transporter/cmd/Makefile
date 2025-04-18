SUBDIRS = kintone #sabakan dell bmc 

.PHONY: all
all: check-generate test

.PHONY: setup
setup:
	go install github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/onsi/ginkgo/v2/ginkgo

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: test
test: clean
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(custom-checker -restrictpkg.packages=html/template ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	#ginkgo -v --race --output-interceptor-mode=none -p .
	ginkgo -v --race -p .

.PHONY: clean
clean:
	rm -f ../testdata/output/*
	go clean -fuzzcache
