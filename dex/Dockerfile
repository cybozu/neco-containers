# ref. https://github.com/dexidp/dex/blob/master/Dockerfile
# This image is designed specifically for Argo CD.
# Note: Upstream tools (xx, stager, gomplate, config.docker.yaml, docker-entrypoint)
#       are intentionally excluded since they are unnecessary here.
FROM ghcr.io/cybozu/golang:1.24.11.1_noble@sha256:f6d175da75638dc4ca2553d55a196454b029963aa2b51a6b6e9a1b4c175766fa AS build

ARG DEX_VERSION=2.43.1
ARG DEX_CHECKSUM=f3b97c0315cbade7072a2665490db2e8cb1869d606f47a63301fd1c5bf568179

WORKDIR /work/dex/
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ADD --checksum=sha256:${DEX_CHECKSUM} \
    https://github.com/dexidp/dex/archive/v${DEX_VERSION}.tar.gz /tmp/dex.tar.gz

RUN tar -x -z --strip-components 1 -f /tmp/dex.tar.gz -C /work/dex 

ENV CGO_ENABLED=1

RUN make release-binary VERSION=v${DEX_VERSION}

FROM ghcr.io/cybozu/ubuntu:24.04.20260508@sha256:ab2735f6893fc167776587097de968c7d66c1cb052326d8cfeb3c4f8cd8bac00
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/dex /usr/local/bin/dex
COPY --from=build /work/dex/LICENSE /usr/local/share/doc/dex/LICENSE
COPY --from=build /work/dex/web /web

USER 10000:10000
