# opentelemetry-collector

FROM quay.io/cybozu/golang:1.20-jammy AS build

ARG OTELCOL_VERSION=0.81.0

WORKDIR /work
COPY manifest.yaml manifest.yaml.in
RUN sed -e s/\${OTELCOL_VERSION}/${OTELCOL_VERSION}/g manifest.yaml.in > manifest.yaml && \
    go install go.opentelemetry.io/collector/cmd/builder@v${OTELCOL_VERSION} && \
    builder --config manifest.yaml
RUN curl -sSf -o LICENSE https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/v${OTELCOL_VERSION}/LICENSE

FROM quay.io/cybozu/ubuntu:22.04

COPY --from=build /work/_build/otelcol /otelcol
COPY --from=build /work/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/otelcol"]
# this container image does not contain default config file.
