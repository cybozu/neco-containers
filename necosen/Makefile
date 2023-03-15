include Makefile.versions

TOOLS_DIR := $(CURDIR)/bin
GH := $(TOOLS_DIR)/gh

##@ Basic

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: setup
setup: ## Setup tools
setup: \
	$(TOOLS_DIR)/gh \
	$(TOOLS_DIR)/helm \
	$(TOOLS_DIR)/kind \
	$(TOOLS_DIR)/kubectl \
	$(TOOLS_DIR)/yq

.PHONY: clean
clean: ## Clean files
	rm -rf bin
	rm -rf upstream

$(TOOLS_DIR)/gh:
	mkdir -p $(TOOLS_DIR)
	curl -fsL https://github.com/cli/cli/releases/download/v$(GH_VERSION)/gh_$(GH_VERSION)_linux_amd64.tar.gz | tar -zx -O gh_$(GH_VERSION)_linux_amd64/bin/gh > $@
	chmod +x $@

$(TOOLS_DIR)/helm:
	mkdir -p $(TOOLS_DIR)
	curl -fsL https://get.helm.sh/helm-v$(HELM_VERSION)-linux-amd64.tar.gz | tar zx -O linux-amd64/helm > $@
	chmod +x $@

$(TOOLS_DIR)/kind:
	GOBIN=$(TOOLS_DIR) go install sigs.k8s.io/kind@v$(KIND_VERSION)

$(TOOLS_DIR)/kubectl:
	mkdir -p $(TOOLS_DIR)
	curl -fsL -o $@ https://storage.googleapis.com/kubernetes-release/release/v$(KUBECTL_VERSION)/bin/linux/amd64/kubectl
	chmod +x $@

$(TOOLS_DIR)/yq:
	mkdir -p $(TOOLS_DIR)
	curl -fsL -o $@ https://github.com/mikefarah/yq/releases/download/v$(YQ_VERSION)/yq_linux_amd64
	chmod +x $@

##@ Maintenance

.PHONY: login-gh
login-gh: ## Login to GitHub
	if ! $(GH) auth status 2>/dev/null; then \
		echo; \
		echo '!! You need login to GitHub to proceed. Please follow the next command with "Authenticate Git with your GitHub credentials? (Y)".'; \
		echo; \
		$(GH) auth login -h github.com -p HTTPS -w; \
	fi

.PHONY: logout-gh
logout-gh: ## Logout from GitHub
	$(GH) auth logout

.PHONY: update
update: ## Update dependent software versions
	$(call get-latest-gh,cli/cli)
	NEW_VERSION=$$(echo $(latest_gh) | cut -b 2-); \
	sed -i -e "s/GH_VERSION := .*/GH_VERSION := $${NEW_VERSION}/g" Makefile.versions

	$(call get-latest-gh,helm/helm)
	NEW_VERSION=$$(echo $(latest_gh) | cut -b 2-); \
	sed -i -e "s/HELM_VERSION := .*/HELM_VERSION := $${NEW_VERSION}/g" Makefile.versions

	$(call get-latest-gh,kubernetes-sigs/kind)
	NEW_VERSION=$$(echo $(latest_gh) | cut -b 2-); \
	sed -i -e "s/KIND_VERSION := .*/KIND_VERSION := $${NEW_VERSION}/g" Makefile.versions

	$(call get-latest-gh,kubernetes/kubernetes)
	NEW_VERSION=$$(echo $(latest_gh) | cut -b 2-); \
	sed -i -e "s/KUBECTL_VERSION := .*/KUBECTL_VERSION := $${NEW_VERSION}/g" Makefile.versions

	$(call get-latest-gh,mikefarah/yq)
	NEW_VERSION=$$(echo $(latest_gh) | cut -b 2-); \
	sed -i -e "s/YQ_VERSION := .*/YQ_VERSION := $${NEW_VERSION}/g" Makefile.versions

##@ Development

.PHONY: build
build: ## Build binary
	CGO_ENABLED=0 go build -o bin/necosen -ldflags="-w -s" main.go

.PHONY:check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: test
test: ## Run e2e test
test: setup
	$(MAKE) -C e2e setup
	$(MAKE) -C e2e start
	$(MAKE) -C e2e test
	$(MAKE) -C e2e stop

# usage get-latest-gh OWNER/REPO
define get-latest-gh
$(eval latest_gh := $(shell $(GH) release list --repo $1 | grep Latest | cut -f3))
endef
