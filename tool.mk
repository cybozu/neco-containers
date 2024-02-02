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

# Download staticcheck locally if necessary
STATICCHECK := $(SELF_DIR)/bin/staticcheck
.PHONY: staticcheck
staticcheck:
	$(call go-install-tool,$(STATICCHECK),honnef.co/go/tools/cmd/staticcheck@latest)

# Download custom-checker locally if necessary
CUSTOM_CHECKER := $(SELF_DIR)/bin/custom-checker
.PHONY: custom-checker
custom-checker:
	$(call go-install-tool,$(CUSTOM_CHECKER),github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest)
