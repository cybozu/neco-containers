# vault container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.10-bionic AS build

ARG VAULT_VERSION=0.10.4

RUN mkdir -p /go/src/github.com/hashicorp/ && \
    git clone --depth=1 -b v${VAULT_VERSION} https://github.com/hashicorp/vault.git /go/src/github.com/hashicorp/vault

WORKDIR /go/src/github.com/hashicorp/vault
RUN make bootstrap && \
    make

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/src/github.com/hashicorp/vault/bin /usr/local/vault/bin
COPY --from=build /go/src/github.com/hashicorp/vault/LICENSE /usr/local/vault/LICENSE

ENV PATH=/usr/local/vault/bin:"$PATH"

EXPOSE 8200

RUN mkdir -p /vault/config

ENTRYPOINT ["/usr/local/vault/bin/vault"]
