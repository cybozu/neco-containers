IMAGE_TAG ?= unbound:dev

.PHONY: update-root-hints
update-root-hints:
	# named.cache file is provided by InterNIC.
	# See: https://jprs.jp/tech/notice/2023-11-28-b.root-servers.net-ip-address-change.html

	# named.cache does not have a LF at the end of file
	curl -so root.hints https://www.internic.net/domain/named.cache
	echo >> root.hints

.PHONY: check-generate
check-generate: update-root-hints
	git diff --exit-code --name-only

.PHONY: validate
validate:
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c "cat /usr/local/unbound/etc/unbound/root.hints"
