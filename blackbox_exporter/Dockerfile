# blackbox-exporter container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.18-focal AS build

ARG BLACKBOX_EXPORTER_VERSION=0.22.0
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
WORKDIR /blackbox_exporter
RUN curl -fsSL https://github.com/prometheus/blackbox_exporter/archive/v${BLACKBOX_EXPORTER_VERSION}.tar.gz | \
    tar -xzf - --strip-components 1

RUN make build

# Stage2: setup runtime container
FROM scratch
COPY --from=build /blackbox_exporter/blackbox_exporter /blackbox_exporter
COPY --from=build /blackbox_exporter/blackbox.yml /blackbox.yml
COPY --from=build /blackbox_exporter/LICENSE /LICENSE

USER 10000:10000
EXPOSE 9115
ENTRYPOINT ["/blackbox_exporter"]