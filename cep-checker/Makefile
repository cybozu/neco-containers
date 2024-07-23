
.PHONEY build:
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o cep-checker .

.PHONEY test:
test:
	$(MAKE) -C e2e setup
	$(MAKE) -C e2e start
	$(MAKE) -C e2e install-cilium
	$(MAKE) -C e2e test
	$(MAKE) -C e2e stop

.PHONEY docker-build:
docker-build:
	docker build -t ghcr.io/cybozu/cep-checker:dev .
