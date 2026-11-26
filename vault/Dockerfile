# vault container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.4.2_resolute@sha256:0ddbebb97d9fefa0961dd10fdccedc40a4b805d28f21f21d92faa2f7b071d2f1 AS build
ENV GOPROXY=https://golang.flatt.tech

ARG VAULT_COMMIT=c56e7e670bd8c79104c27b6c10138226538a6db2 # v2.2.2

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    git clone --depth 1 https://github.com/hashicorp/vault . \
    && git fetch --depth 1 origin ${VAULT_COMMIT} \
    && git checkout FETCH_HEAD \
    && make bootstrap \
    && make

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:26.04.20261126@sha256:c9ec282d212c595558d8dbb1cc75a0dc9d82174979bcb724ca116f3695cc6022
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/bin /usr/local/vault/bin
COPY --from=build /work/LICENSE /usr/local/vault/LICENSE
COPY install-tools /usr/local/vault/install-tools

ENV PATH=/usr/local/vault/bin:"$PATH"

USER 10000:10000
EXPOSE 8200 8201

ENTRYPOINT ["/usr/local/vault/bin/vault"]
