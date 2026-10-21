# ref. https://github.com/dexidp/dex/blob/master/Dockerfile
# This image is designed specifically for Argo CD.
# Note: Upstream tools (xx, stager, gomplate, config.docker.yaml, docker-entrypoint)
#       are intentionally excluded since they are unnecessary here.
FROM ghcr.io/cybozu/golang:1.26.10.2_noble@sha256:8d900a9ee78c02dbe4e2d2e23800c4d21714dfaf8c679935afa668ca0993a0e4 AS build
ENV GOPROXY=https://golang.flatt.tech

ARG DEX_VERSION=2.45.1

WORKDIR /work/dex/
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ADD --checksum=sha256:6dc9bf768d67723d8117268d5177c7450519608f5403cd10e857c89465fa2b9e \
    https://github.com/dexidp/dex/archive/v${DEX_VERSION}.tar.gz /tmp/dex.tar.gz

RUN tar -x -z --strip-components 1 -f /tmp/dex.tar.gz -C /work/dex 

ENV CGO_ENABLED=1

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    make release-binary VERSION=v${DEX_VERSION}

FROM ghcr.io/cybozu/ubuntu:24.04.20261021@sha256:7bc89f8e6e4dfe609ca671526e3bbee91753422b8a766d90cef4de8b7c73a3b9
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/dex /usr/local/bin/dex
COPY --from=build /work/dex/LICENSE /usr/local/share/doc/dex/LICENSE
COPY --from=build /work/dex/web /web

USER 10000:10000
