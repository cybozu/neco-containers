TRUST_MANAGER_VERSION := $(shell cat TAG | cut -d'.' -f1-3)
TRUST_MANAGER_SOURCE := trust-manager-source

.PHONY: clean
clean:
	rm -rf ./$(TRUST_MANAGER_SOURCE)

$(TRUST_MANAGER_SOURCE):
	mkdir $@
	cd $@; \
	curl -fsSL https://github.com/cert-manager/trust-manager/archive/v$(TRUST_MANAGER_VERSION).tar.gz | \
    tar --strip-components=1 -xzf -

$(TRUST_MANAGER_SOURCE)/validate-trust-package: | $(TRUST_MANAGER_SOURCE)
	cd $(TRUST_MANAGER_SOURCE); go build -o validate-trust-package ./cmd/validate-trust-package/main.go

# debian-trust-package-fetch.sh uses docker command and run a specified container image to write package.json,
# that is the bundle of Certificates of the common public CAs retrieved from the ca-certificates package.
# ref: https://github.com/cert-manager/trust-manager/blob/main/make/debian-trust-package-fetch.sh
# It is hard to embed the work into Dockerfile, as it uses docker command.
$(TRUST_MANAGER_SOURCE)/package.json: | $(TRUST_MANAGER_SOURCE) $(TRUST_MANAGER_SOURCE)/validate-trust-package
	BIN_VALIDATE_TRUST_PACKAGE=$(TRUST_MANAGER_SOURCE)/validate-trust-package \
	$(TRUST_MANAGER_SOURCE)/make/debian-trust-package-fetch.sh latest ghcr.io/cybozu/ubuntu:22.04 \
	$(TRUST_MANAGER_SOURCE)/package.json latest cert-manager-debian

.PHONY: download-certificates
download-certificates: | $(TRUST_MANAGER_SOURCE)/package.json
