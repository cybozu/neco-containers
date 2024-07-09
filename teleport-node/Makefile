.PHONY: check-teleport-update
check-teleport-update:
	rm -rf /tmp/teleport
	mkdir /tmp/teleport
	$(call get-latest-gh-package-tag,cybozu/teleport-node)
	$(call upstream-tag,$(latest_tag))
	$(call get-latest-gh,gravitational/teleport)
	echo $(latest_gh)

	mkdir /tmp/teleport/$(upstream_tag)
	cd /tmp/teleport/$(upstream_tag) && \
	curl -sSLf -o Makefile https://raw.githubusercontent.com/gravitational/teleport/$(upstream_tag)/Makefile && \
	curl -sSLf -o versions.mk https://raw.githubusercontent.com/gravitational/teleport/$(upstream_tag)/build.assets/versions.mk

	mkdir /tmp/teleport/$(latest_gh)
	cd /tmp/teleport/$(latest_gh) && \
	curl -sSLf -o Makefile https://raw.githubusercontent.com/gravitational/teleport/$(latest_gh)/Makefile && \
	curl -sSLf -o versions.mk https://raw.githubusercontent.com/gravitational/teleport/$(latest_gh)/build.assets/versions.mk

	diff -u /tmp/teleport/$(upstream_tag) /tmp/teleport/$(latest_gh)

# usage: get-latest-gh-package-tag ORG/NAME
define get-latest-gh-package-tag
$(eval latest_tag := $(shell curl -sSf -H "Authorization: Bearer $(shell curl -sSf "https://ghcr.io/token?scope=repository:$(shell echo $1 | sed 's/\//%2F/g')%3Apull&service=ghcr.io" | jq -r .token)" https://ghcr.io/v2/$(shell echo $1 | sed 's/\//%2F/g')/tags/list | jq -r '.tags[]' | grep -v latest | sort -Vr | head -n 1))
endef

# usage: upstream-tag 1.2.3.4
define upstream-tag
$(eval upstream_tag := $(shell echo $1 | sed -E 's/^(.*)\.[[:digit:]]+$$/v\1/'))
endef

# usage get-latest-gh OWNER/REPO
define get-latest-gh
$(eval latest_gh := $(shell curl -sSf https://api.github.com/repos/$1/releases/latest | jq -r '.tag_name'))
endef
