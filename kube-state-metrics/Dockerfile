# kube-state-metrics container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.4.2_resolute@sha256:0ddbebb97d9fefa0961dd10fdccedc40a4b805d28f21f21d92faa2f7b071d2f1 AS build
ENV GOPROXY=https://golang.flatt.tech

# https://github.com/kubernetes/kube-state-metrics/releases/tag/v2.20.0
ARG KUBE_STATE_METRICS_COMMIT=4ffeda2ef866b0fef372849825a803296483b336
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/kubernetes/kube-state-metrics/archive/${KUBE_STATE_METRICS_COMMIT}.tar.gz | \
    tar --strip-components=1 -xzf -

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    make build-local

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/kube-state-metrics /kube-state-metrics
COPY --from=build /work/LICENSE /LICENSE
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

USER 10000:10000
EXPOSE 8080 8081
ENTRYPOINT ["/kube-state-metrics", "--port=8080", "--telemetry-port=8081"]
