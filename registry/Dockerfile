# Registry container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG REGISTRY_VERSION=3.0.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p /go/src/github.com/docker/ \
    && curl -sSLf https://github.com/distribution/distribution/archive/v${REGISTRY_VERSION}.tar.gz | \
    tar zxf - -C /go/src/github.com/docker/ \
    && mv /go/src/github.com/docker/distribution-${REGISTRY_VERSION} /go/src/github.com/docker/distribution

WORKDIR /go/src/github.com/docker/distribution/

RUN CGO_ENABLED=0 make clean binaries VERSION=${REGISTRY_VERSION} REVISION=""

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update && apt-get install -y --no-install-recommends \
    jq && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /go/src/github.com/docker/distribution/bin/registry /bin/registry
COPY --from=build /go/src/github.com/docker/distribution/LICENSE /usr/local/distribution/LICENSE
COPY --from=build /go/src/github.com/docker/distribution/cmd/registry/config-dev.yml /etc/distribution/config.yml

USER 10000:10000
VOLUME ["/var/lib/registry"]
EXPOSE 5000

ENTRYPOINT ["registry"]
CMD ["serve", "/etc/distribution/config.yml"]
