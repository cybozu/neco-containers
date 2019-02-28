# vault container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG VAULT_VERSION=1.0.3

RUN mkdir -p /go/src/github.com/hashicorp/ && \
    git clone --depth=1 -b v${VAULT_VERSION} \
        https://github.com/hashicorp/vault.git /go/src/github.com/hashicorp/vault

WORKDIR /go/src/github.com/hashicorp/vault

COPY disable-etcd-version-detection.patch .
RUN patch -p1 < disable-etcd-version-detection.patch

RUN make fmt && \
    make bootstrap && \
    make

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/src/github.com/hashicorp/vault/bin /usr/local/vault/bin
COPY --from=build /go/src/github.com/hashicorp/vault/LICENSE /usr/local/vault/LICENSE
COPY install-tools /usr/local/vault/install-tools

ENV PATH=/usr/local/vault/bin:"$PATH"

USER 10000:10000
EXPOSE 8200 8201

ENTRYPOINT ["/usr/local/vault/bin/vault"]
