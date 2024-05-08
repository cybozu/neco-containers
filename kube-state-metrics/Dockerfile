# kube-state-metrics container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG KUBE_STATE_METRICS_VERSION=2.12.0
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/kubernetes/kube-state-metrics/archive/v${KUBE_STATE_METRICS_VERSION}.tar.gz | \
    tar --strip-components=1 -xzf -

RUN make build-local

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/kube-state-metrics /kube-state-metrics
COPY --from=build /work/LICENSE /LICENSE

USER 10000:10000
EXPOSE 8080 8081
ENTRYPOINT ["/kube-state-metrics", "--port=8080", "--telemetry-port=8081"]
