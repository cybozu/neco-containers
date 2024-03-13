# vault container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG VAULT_VERSION=1.15.6

RUN git clone --depth=1 -b v${VAULT_VERSION} https://github.com/hashicorp/vault.git

WORKDIR /work/vault

RUN make bootstrap && \
    make fmt && \
    make

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/vault/bin /usr/local/vault/bin
COPY --from=build /work/vault/LICENSE /usr/local/vault/LICENSE
COPY install-tools /usr/local/vault/install-tools

ENV PATH=/usr/local/vault/bin:"$PATH"

USER 10000:10000
EXPOSE 8200 8201

ENTRYPOINT ["/usr/local/vault/bin/vault"]
