IMAGE_NAME=ghcr.io/cybozu-private/bucket-provisioner-light:dev

.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	mkdir -p ./bin
	go build -o ./bin/bucket-provisioner-light main.go

.PHONY: build-image
build-image:
	docker build -t $(IMAGE_NAME) .
