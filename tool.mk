define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(dir $(1)) go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

SELF_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

# Download goimports locally if necessary
GOIMPORTS := $(SELF_DIR)/bin/goimports
.PHONY: goimports
goimports:
	$(call go-install-tool,$(GOIMPORTS),golang.org/x/tools/cmd/goimports@latest)

# Download staticcheck locally if necessary
STATICCHECK := $(SELF_DIR)/bin/staticcheck
.PHONY: staticcheck
staticcheck:
	$(call go-install-tool,$(STATICCHECK),honnef.co/go/tools/cmd/staticcheck@latest)

# Download buf locally if necessary
BUF := $(SELF_DIR)/bin/buf
.PHONY: buf
buf:
	$(call go-install-tool,$(BUF),github.com/bufbuild/buf/cmd/buf@873c86f6e17c9e9d9a747ea1c521f4ed580ab5d7) # v1.66.1

# Download gofumpt locally if necessary
GOFUMPT := $(SELF_DIR)/bin/gofumpt
.PHONY: gofumpt
gofumpt:
	$(call go-install-tool,$(GOFUMPT),mvdan.cc/gofumpt@718975501de6321ddf0a5fd17b4f959d33fa203e) # v0.9.2

# Download custom-checker locally if necessary
CUSTOM_CHECKER := $(SELF_DIR)/bin/custom-checker
.PHONY: custom-checker
custom-checker:
	$(call go-install-tool,$(CUSTOM_CHECKER),github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest)
