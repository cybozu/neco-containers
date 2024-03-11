# blackbox-exporter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG BLACKBOX_EXPORTER_VERSION=0.24.0
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
WORKDIR /blackbox_exporter
RUN curl -fsSL https://github.com/prometheus/blackbox_exporter/archive/v${BLACKBOX_EXPORTER_VERSION}.tar.gz | \
    tar -xzf - --strip-components 1

RUN make build

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /blackbox_exporter/blackbox_exporter /blackbox_exporter
COPY --from=build /blackbox_exporter/blackbox.yml /blackbox.yml
COPY --from=build /blackbox_exporter/LICENSE /LICENSE

USER 10000:10000
EXPOSE 9115
ENTRYPOINT ["/blackbox_exporter"]
