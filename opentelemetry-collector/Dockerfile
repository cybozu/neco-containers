# opentelemetry-collector

FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG OTELCOL_VERSION=0.129.0
ARG OTELCOL_CONTRIB_VERSION=0.129.0

WORKDIR /work
COPY manifest.yaml manifest.yaml.in
RUN cat manifest.yaml.in | \
        sed -e s/\${OTELCOL_VERSION}/${OTELCOL_VERSION}/g | \
        sed -e s/\${OTELCOL_CONTRIB_VERSION}/${OTELCOL_CONTRIB_VERSION}/g > manifest.yaml && \
    go install go.opentelemetry.io/collector/cmd/builder@v${OTELCOL_VERSION} && \
    builder --config manifest.yaml
RUN curl -sSf -o LICENSE https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/v${OTELCOL_VERSION}/LICENSE

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/_build/otelcol /otelcol
COPY --from=build /work/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/otelcol"]
# this container image does not contain default config file.
