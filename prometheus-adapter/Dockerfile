# prometheus-adapter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG PROMETHEUS_ADAPTER_COMMIT=17cef511b1854441490bceeca7a710a04ce091ad

RUN curl -fsSL -o pa.tar.gz https://github.com/kubernetes-sigs/prometheus-adapter/archive/${PROMETHEUS_ADAPTER_COMMIT}.tar.gz \
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
