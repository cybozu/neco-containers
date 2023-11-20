SUDO = sudo

.PHONY: all
all: check-generate test

.PHONY: dependencies
dependencies:
	GOBIN=/usr/local/bin go install honnef.co/go/tools/cmd/staticcheck@latest
	GOBIN=/usr/local/bin go install github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: test
test: dependencies
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	go vet ./...
	go test -c ./...
	$(SUDO) ./bpf-map-pressure-exporter.test -test.v
	rm -f ./bpf-map-pressure-exporter.test
