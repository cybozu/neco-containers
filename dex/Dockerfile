# ref. https://github.com/dexidp/dex/blob/master/Dockerfile
# This image is designed specifically for Argo CD.
# Note: Upstream tools (xx, stager, gomplate, config.docker.yaml, docker-entrypoint)
#       are intentionally excluded since they are unnecessary here.
FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG DEX_VERSION=2.43.1

WORKDIR /work/dex/
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN curl -sSLf https://github.com/dexidp/dex/archive/v${DEX_VERSION}.tar.gz | \
    tar zxf - -C /work/dex --strip-components 1

ENV CGO_ENABLED=1

RUN make release-binary VERSION=v${DEX_VERSION}

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/dex /usr/local/bin/dex
COPY --from=build /work/dex/LICENSE /usr/local/share/doc/dex/LICENSE
COPY --from=build /work/dex/web /web

USER 10000:10000
