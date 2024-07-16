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
	rm -f pointers/*
	rm -f pointers_it1/*
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	ginkgo -procs=1 --race --focus-file ut_get_machine-list_test.go
	ginkgo -procs=1 --race --focus-file ut_access_idrac_test.go
	ginkgo -procs=1 --race --focus-file it1_get_logs_test.go
	#ginkgo -procs=1 --focus-file it2_mainloop_test.go
