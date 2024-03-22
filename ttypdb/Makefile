.PHONY: all
all:

.PHONY: start-kind
start-kind:
	kind create cluster

.PHONY: stop-kind
stop-kind:
	kind delete cluster

.PHONY: e2e
e2e:
	cd src && go test ./... -run TestE2E

.PHONY: update-image
update-image: docker
	kind load docker-image cybozu/ttypdb-sidecar:dev
	kind load docker-image cybozu/ttypdb-controller:dev

.PHONY: docker
docker:
	docker build . -t cybozu/ttypdb-sidecar:dev --target sidecar
	docker build . -t cybozu/ttypdb-controller:dev --target controller
