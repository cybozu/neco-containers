IMAGE_TAG ?= ghcr.io/cybozu/neco-server-exporter:$(shell cat TAG)

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY build:
build:
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/neco-server-exporter .

.PHONY: test
test: staticcheck custom-checker
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	$(STATICCHECK) ./...
	test -z "$$($(CUSTOM_CHECKER) -restrictpkg.packages=html/template,log ./... 2>&1 | tee /dev/stderr)"
	go vet ./...

.PHONY: docker-build
docker-build:
	docker build . --tag=$(IMAGE_TAG)

.PHONY: sanity-check
sanity-check:
	docker run --rm $(IMAGE_TAG) --help

.PHONY: install
install: docker-build
	kind load docker-image $(IMAGE_TAG)
	kustomize build testdata | kubectl apply -f -

.PHONY: uninstall
uninstall:
	kustomize build testdata | kubectl delete -f -

include ../tool.mk
