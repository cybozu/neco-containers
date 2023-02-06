# Registry container image

# Stage1: build from source
FROM quay.io/cybozu/golang:1.19-jammy AS build

ARG REGISTRY_VERSION=2.8.1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p /go/src/github.com/docker/ \
    && curl -sSLf https://github.com/distribution/distribution/archive/v${REGISTRY_VERSION}.tar.gz | \
        tar zxf - -C /go/src/github.com/docker/ \
    && mv /go/src/github.com/docker/distribution-${REGISTRY_VERSION} /go/src/github.com/docker/distribution

WORKDIR /go/src/github.com/docker/distribution/

RUN GO111MODULE=off CGO_ENABLED=0 make clean binaries VERSION=${REGISTRY_VERSION} REVISION=""

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:22.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    jq && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /go/src/github.com/docker/distribution/bin/registry /bin/registry
COPY --from=build /go/src/github.com/docker/distribution/LICENSE /usr/local/distribution/LICENSE
COPY --from=build /go/src/github.com/docker/distribution/cmd/registry/config-dev.yml /etc/docker/registry/config.yml

USER 10000:10000
VOLUME ["/var/lib/registry"]
EXPOSE 5000

ENTRYPOINT ["registry"]
CMD ["serve", "/etc/docker/registry/config.yml"]
