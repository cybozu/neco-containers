# Registry container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.4.2_resolute@sha256:0ddbebb97d9fefa0961dd10fdccedc40a4b805d28f21f21d92faa2f7b071d2f1 AS build

ARG REGISTRY_COMMIT=3220848f15d9279c66a41aa0a257469e6aede1e9 # v3.1.2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p /go/src/github.com/distribution/ \
    && curl -sSLf https://github.com/distribution/distribution/archive/${REGISTRY_COMMIT}.tar.gz | \
    tar zxf - -C /go/src/github.com/distribution/ \
    && mv /go/src/github.com/distribution/distribution-${REGISTRY_COMMIT} /go/src/github.com/distribution/distribution

WORKDIR /go/src/github.com/distribution/distribution/

RUN CGO_ENABLED=0 make clean binaries

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:26.04.20261117@sha256:47ff7a519b04321a84301061371358e06e97f80f4ef0229e5e2ad2e7f71acb55
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update && apt-get install -y --no-install-recommends \
    jq && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /go/src/github.com/distribution/distribution/bin/registry /bin/registry
COPY --from=build /go/src/github.com/distribution/distribution/LICENSE /usr/local/distribution/LICENSE
COPY --from=build /go/src/github.com/distribution/distribution/cmd/registry/config-dev.yml /etc/distribution/config.yml

USER 10000:10000
VOLUME ["/var/lib/registry"]
EXPOSE 5000

ENTRYPOINT ["registry"]
CMD ["serve", "/etc/distribution/config.yml"]
