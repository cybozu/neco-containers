TRUST_MANAGER_VERSION := $(shell cat TAG | cut -d'.' -f1-3)
TRUST_MANAGER_HASH := c1ba3525637adc858a8addb24b87be4d0189a77c0e532734d9aeac41b3351f71
TRUST_MANAGER_SOURCE := trust-manager-source
UBUNTU_IMAGE := ghcr.io/cybozu/ubuntu:24.04.20260709@sha256:f91f1884f5b8dbe44e224306b5dc7bfd4d420f5c73aea1d32a71c6474361fb0f

.PHONY: clean
clean:
	rm -rf ./$(TRUST_MANAGER_SOURCE)

$(TRUST_MANAGER_SOURCE):
	$(eval TMPDIR := $(shell mktemp -d))
	echo $(TMPDIR)
	mkdir $@
	curl -fsSL https://github.com/cert-manager/trust-manager/archive/v$(TRUST_MANAGER_VERSION).tar.gz -o $(TMPDIR)/trust-manager.tar.gz
	echo "$(TRUST_MANAGER_HASH) $(TMPDIR)/trust-manager.tar.gz" | sha256sum -c -
	tar zxf $(TMPDIR)/trust-manager.tar.gz --strip-components=1 -C $@
	rm -rf $(TMPDIR)

$(TRUST_MANAGER_SOURCE)/validate-trust-package: | $(TRUST_MANAGER_SOURCE)
	cd $(TRUST_MANAGER_SOURCE); go build -o validate-trust-package ./cmd/validate-trust-package/main.go

# debian-trust-package-fetch.sh uses docker command and run a specified container image to write package.json,
# that is the bundle of Certificates of the common public CAs retrieved from the ca-certificates package.
# ref: https://github.com/cert-manager/trust-manager/blob/main/make/debian-trust-package-fetch.sh
# It is hard to embed the work into Dockerfile, as it uses docker command.
$(TRUST_MANAGER_SOURCE)/package.json: | $(TRUST_MANAGER_SOURCE) $(TRUST_MANAGER_SOURCE)/validate-trust-package
	BIN_VALIDATE_TRUST_PACKAGE=$(TRUST_MANAGER_SOURCE)/validate-trust-package \
	$(TRUST_MANAGER_SOURCE)/make/debian-trust-package-fetch.sh latest $(UBUNTU_IMAGE) \
	$(TRUST_MANAGER_SOURCE)/package.json latest cert-manager-debian

.PHONY: download-certificates
download-certificates: | $(TRUST_MANAGER_SOURCE)/package.json
