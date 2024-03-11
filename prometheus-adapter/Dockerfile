# prometheus-adapter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG PROMETHEUS_ADAPTER_VERSION=0.11.2

RUN curl -fsSL -o pa.tar.gz https://github.com/kubernetes-sigs/prometheus-adapter/archive/v${PROMETHEUS_ADAPTER_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f pa.tar.gz \
    && rm -f pa.tar.gz \
    && make all

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/adapter /adapter
COPY --from=build /work/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/adapter"]
