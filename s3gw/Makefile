SUDO = sudo
SUDO_GO = $(SUDO) $(shell which go)

.PHONY: all
all: check-generate test

.PHONY: setup
setup:
	go install github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@5cda2f85e31dbe2453825f6520710a76465f197e # v0.1.5
	go install honnef.co/go/tools/cmd/staticcheck@ff63afafc529279f454e02f1d060210bd4263951 # v0.7.0

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
	$(SUDO_GO) test -race -v ./...
