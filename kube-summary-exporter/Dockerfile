# kube-summary-exporter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.20-jammy AS build

ARG KUBE_SUMMARY_EXPORTER_VERSION=0.4.2
WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/utilitywarehouse/kube-summary-exporter/archive/v${KUBE_SUMMARY_EXPORTER_VERSION}.tar.gz | \
    tar --strip-components=1 -xzf -

RUN CGO_ENABLED=0 go build -o /kube-summary-exporter

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers/kube-summary-exporter"

COPY --from=build /kube-summary-exporter /kube-summary-exporter
COPY --from=build /work/LICENSE /LICENSE

USER 10000:10000
EXPOSE 9779 9779
ENTRYPOINT ["/kube-summary-exporter", "--listen-address=:9779"]
