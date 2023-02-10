# kube-storage-version-migrator containers
# This Dockerfile builds 3 final targets:
#   1. initializer
#   2. migrator
#   3. trigger
# Each target is prefixed with "storage-version-migrate-" to construct its repository name.

# Stage1: build from source
FROM quay.io/cybozu/golang:1.19-jammy AS build

ARG MIGRATOR_VERSION=0.0.5

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/kubernetes-sigs
RUN git clone --depth 1 --branch v${MIGRATOR_VERSION} https://github.com/kubernetes-sigs/kube-storage-version-migrator \
    && cd kube-storage-version-migrator \
    && mkdir -p /work \
    && CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -mod=vendor -ldflags "-X sigs.k8s.io/kube-storage-version-migrator/pkg/version.VERSION=v${MIGRATOR_VERSION}" -a -installsuffix cgo -o /work/ ./cmd/...

# Stage2 for initializer: setup runtime container
FROM scratch AS initializer

COPY --from=build /work/initializer /initializer
USER 10000:10000
ENTRYPOINT ["/initializer", "--alsologtostderr", "--v=2"]

# Stage2 for migrator: setup runtime container
FROM scratch AS migrator

COPY --from=build /work/migrator /migrator
USER 10000:10000
ENTRYPOINT ["/migrator", "--alsologtostderr", "--v=2"]

# Stage2 for trigger: setup runtime container
FROM scratch AS trigger

COPY --from=build /work/trigger /trigger
USER 10000:10000
ENTRYPOINT ["/trigger", "--alsologtostderr", "--v=2"]
