# vault container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build

ARG VAULT_VERSION=1.4.1

RUN git clone --depth=1 -b v${VAULT_VERSION} https://github.com/hashicorp/vault.git

WORKDIR /work/vault

COPY disable-etcd-version-detection.patch .
COPY remove-broken-golanci-lint.patch .
RUN patch -p1 < disable-etcd-version-detection.patch
RUN patch -p1 < remove-broken-golanci-lint.patch

RUN make fmt && \
    make bootstrap && \
    GOFLAGS=-mod=vendor make

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /work/vault/bin /usr/local/vault/bin
COPY --from=build /work/vault/LICENSE /usr/local/vault/LICENSE
COPY install-tools /usr/local/vault/install-tools

ENV PATH=/usr/local/vault/bin:"$PATH"

USER 10000:10000
EXPOSE 8200 8201

ENTRYPOINT ["/usr/local/vault/bin/vault"]
