# ref. https://github.com/dexidp/dex/blob/master/Dockerfile
# This image is designed specifically for Argo CD.
# Note: Upstream tools (xx, stager, gomplate, config.docker.yaml, docker-entrypoint)
#       are intentionally excluded since they are unnecessary here.
FROM ghcr.io/cybozu/golang:1.26.4.1_noble@sha256:add9d704d4b75df2c51328615be89b61a3e71e4833321aa02c3f325a30d3eb8f AS build

ARG DEX_VERSION=2.45.1

WORKDIR /work/dex/
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ADD --checksum=sha256:6dc9bf768d67723d8117268d5177c7450519608f5403cd10e857c89465fa2b9e \
    https://github.com/dexidp/dex/archive/v${DEX_VERSION}.tar.gz /tmp/dex.tar.gz

RUN tar -x -z --strip-components 1 -f /tmp/dex.tar.gz -C /work/dex 

ENV CGO_ENABLED=1

RUN make release-binary VERSION=v${DEX_VERSION}

FROM ghcr.io/cybozu/ubuntu:24.04.20260604@sha256:c7c519ae7504df61e01495635ebda9c8698f2d221cd3fcf6e9500ad97b9d2421
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/dex /usr/local/bin/dex
COPY --from=build /work/dex/LICENSE /usr/local/share/doc/dex/LICENSE
COPY --from=build /work/dex/web /web

USER 10000:10000
