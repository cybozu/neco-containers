# Registry container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG REGISTRY_COMMIT=9a8d98b679740cd514aa7e7d84d23d442a5ef54c

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p /go/src/github.com/distribution/ \
    && curl -sSLf https://github.com/distribution/distribution/archive/${REGISTRY_COMMIT}.tar.gz | \
    tar zxf - -C /go/src/github.com/distribution/ \
    && mv /go/src/github.com/distribution/distribution-${REGISTRY_COMMIT} /go/src/github.com/distribution/distribution

WORKDIR /go/src/github.com/distribution/distribution/

RUN CGO_ENABLED=0 make clean binaries

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04
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
