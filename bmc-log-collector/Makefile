SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: check-generate test

.PHONY: setup
setup:
	go install github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: test
test:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	#ginkgo -v -procs=1 --race --focus-file queue_test.go
	#ginkgo -v -procs=1 --race --focus-file machines-list_test.go
	#ginkgo -v -procs=1 --race --focus-file bmc-http-client_test.go
	#ginkgo -v -procs=1 --race --focus-file log-pointer_test.go
	#ginkgo -v -procs=1 --race --focus-file log-collector1_test.go
	ginkgo -v -procs=1 --race --focus-file log-collector2_test.go
