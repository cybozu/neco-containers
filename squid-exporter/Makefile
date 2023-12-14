
.PHONEY build:
build:
	CGO_ENABLED=0 go build -o squiud-exporter .

.PHONEY run-squid:
run-squid:
	docker run --rm -it --net=host --name squid ghcr.io/cybozu/squid:6

.PHONEY test:
test:
	go test -v ./...

.PHONEY docker-build:
docker-build:
	docker build -t ghcr.io/cybozu/squid-exporter:dev .
