TARGET  = tsr-transporter
TAG = `cat TAG`

.PHONY: all
all: lint build

.PHONY: setup
setup:
	go install github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/onsi/ginkgo/v2/ginkgo

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint: clean
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(custom-checker -restrictpkg.packages=html/template ./... 2>&1 | tee /dev/stderr)"
	go vet ./...

.PHONY: build
build: lint
	go build -o $(TARGET) main.go

.PHONY: image
image: build 
	docker build -t $(TARGET):$(TAG) .

.PHONY: clean
clean:
	rm -f $(TARGET)
	rm -f testdata/output/*
	go clean -fuzzcache

# DO NOT RUN IN CI
# This test needs to setup kintone and iDRAC.
# You should use in development phase.
SUBDIRS = kintone bmc cmd dell sabakan 
test-all:
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir test; \
	done
