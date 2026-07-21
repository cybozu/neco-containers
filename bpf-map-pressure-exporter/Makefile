SUDO = sudo

.PHONY: all
all: lint check-generate test

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint:
	aqua install
	golangci-lint run

.PHONY: test
test:
	go test -c ./...
	$(SUDO) ./bpf-map-pressure-exporter.test -test.v
	rm -f ./bpf-map-pressure-exporter.test

include ../tool.mk
