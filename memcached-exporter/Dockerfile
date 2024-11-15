# memcached_exporter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.23-jammy AS build

ARG MEMCACHED_EXPORTER_VERSION=0.15.0

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN git clone --depth=1 -b v${MEMCACHED_EXPORTER_VERSION} https://github.com/prometheus/memcached_exporter /work/memcached_exporter

WORKDIR /work/memcached_exporter
RUN make build

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/memcached_exporter/memcached_exporter /bin/memcached_exporter
COPY --from=build /work/memcached_exporter/LICENSE /LICENSE

USER 10000:10000
EXPOSE 9150
ENTRYPOINT ["/bin/memcached_exporter"]
