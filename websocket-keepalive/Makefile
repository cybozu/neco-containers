TAG ?= dev

.PHONY: build
build:
	go build -o websocket-keepalive .

.PHONY: image
image:
	docker build -t ghcr.io/cybozu/websocket-keepalive:$(TAG) .

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -f websocket-keepalive
