
.PHONY build:
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o squid-exporter .

.PHONY run-squid:
run-squid:
	docker run --rm -it --net=host --name squid ghcr.io/cybozu/squid:6

.PHONY test:
test:
	go test -v ./...
	make -C e2e setup
	make -C e2e start
	make -C e2e test

.PHONY unit-test:
unit-test:
	go test -v ./...

.PHONY docker-build:
docker-build:
	docker build -t ghcr.io/cybozu/squid-exporter:dev .
