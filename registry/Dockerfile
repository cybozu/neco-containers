# Registry container image

ARG DISTRIBUTION_VERSION=3.1.1

# Stage0: download source archive
FROM scratch AS source

ARG DISTRIBUTION_VERSION
ARG DISTRIBUTION_TARBALL_SHA256=419183e8c1426b5d625732cf6846a2d48ad8e898a1659d7bec7171b7d156673e

ADD --checksum=sha256:${DISTRIBUTION_TARBALL_SHA256} \
    https://github.com/distribution/distribution/archive/v${DISTRIBUTION_VERSION}.tar.gz \
    /distribution.tar.gz

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.25-noble AS build

ARG DISTRIBUTION_VERSION

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
COPY --from=source /distribution.tar.gz distribution.tar.gz
RUN mkdir -p /go/src/github.com/docker/ \
    && tar zxf distribution.tar.gz -C /go/src/github.com/docker/ \
    && mv /go/src/github.com/docker/distribution-${DISTRIBUTION_VERSION} /go/src/github.com/docker/distribution

WORKDIR /go/src/github.com/docker/distribution/

RUN CGO_ENABLED=0 make clean binaries VERSION=${DISTRIBUTION_VERSION} REVISION=""

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
