# vault container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG VAULT_COMMIT=ffe7023c481dc1ea2d8550bbaca8d85f8e611e0b # 1.21.4

RUN git clone --depth 1 https://github.com/hashicorp/vault . \
    && git fetch --depth 1 origin ${VAULT_COMMIT} \
    && git checkout FETCH_HEAD \
    && make bootstrap \
    && make

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/bin /usr/local/vault/bin
COPY --from=build /work/LICENSE /usr/local/vault/LICENSE
COPY install-tools /usr/local/vault/install-tools

ENV PATH=/usr/local/vault/bin:"$PATH"

USER 10000:10000
EXPOSE 8200 8201

ENTRYPOINT ["/usr/local/vault/bin/vault"]
