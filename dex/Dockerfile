# dex container image

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG DEX_VERSION=2.19.0

WORKDIR /work/dex/
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/dexidp/dex/archive/v${DEX_VERSION}.tar.gz | \
        tar zxf - -C /work/dex --strip-components 1

RUN make bin/dex

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        openssl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=build /work/dex/bin/dex /dex
COPY --from=build /work/dex/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/dex"]
