# vault container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.20-jammy AS build

ARG VAULT_VERSION=1.13.0

RUN git clone --depth=1 -b v${VAULT_VERSION} https://github.com/hashicorp/vault.git

WORKDIR /work/vault

RUN make bootstrap && \
    make fmt && \
    make

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:22.04

COPY --from=build /work/vault/bin /usr/local/vault/bin
COPY --from=build /work/vault/LICENSE /usr/local/vault/LICENSE
COPY install-tools /usr/local/vault/install-tools

ENV PATH=/usr/local/vault/bin:"$PATH"

USER 10000:10000
EXPOSE 8200 8201

ENTRYPOINT ["/usr/local/vault/bin/vault"]
