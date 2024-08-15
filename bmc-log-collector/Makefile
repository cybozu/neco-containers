GINKGO = $(HOME)/go/bin/ginkgo

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
test:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(custom-checker -restrictpkg.packages=html/template ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	#$(GINKGO) -v -procs=1 --race --focus-file machines-list_test.go
	#$(GINKGO) -v -procs=1 --race --focus-file bmc-http-client_test.go
	#$(GINKGO) -v -procs=1 --race --focus-file log-pointer_test.go
	#$(GINKGO) -v -procs=1 --race --focus-file log-collector_test.go
	$(GINKGO) -v -procs=1 --race --focus-file main_test.go
	#$(GINKGO) -v -procs=1 --race .
