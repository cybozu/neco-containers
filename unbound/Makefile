IMAGE_TAG ?= unbound:dev

validate:
	docker run --rm --entrypoint bash $(IMAGE_TAG) -c "cat /usr/local/unbound/etc/unbound/root.hints"
