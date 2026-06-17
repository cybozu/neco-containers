REGISTRY ?= ghcr.io/cybozu
IMAGE_NAME ?= sbomreports-github-syncer
TAG ?= dev
IMAGE := $(REGISTRY)/$(IMAGE_NAME):$(TAG)

.PHONY: build
build:
	docker build -t $(IMAGE) .

.PHONY: kind-load
kind-load:
	kind load docker-image $(IMAGE) --name kind-cluster

.PHONY: test
test:
	go test -v ./...
