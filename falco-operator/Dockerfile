FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG FALCO_OPERATOR_COMMIT=ddf43f9fc0f115a34246f6267ca088f7e51b08de # v0.4.1
ARG ARTIFACT_OPERATOR_IMAGE=ghcr.io/cybozu/falco-artifact-operator:0.4.1.1
ARG PROJECT=github.com/falcosecurity/falco-operator

WORKDIR /work/falco-operator
RUN curl -fsSL -o /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz https://github.com/falcosecurity/falco-operator/archive/${FALCO_OPERATOR_COMMIT}.tar.gz && \
    tar -xzf /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz --strip-components=1 && \
    rm /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz && \
    CGO_ENABLED=0 go build -ldflags \
        "-s -w -X '${PROJECT}/internal/pkg/version.ArtifactOperatorImage=${ARTIFACT_OPERATOR_IMAGE}'" \
        -o /usr/local/bin/falco-operator ./cmd/instance/main.go && \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o /usr/local/bin/falco-artifact-operator ./cmd/artifact/main.go

FROM scratch AS falco-operator
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /usr/local/bin/falco-operator /usr/bin/manager
COPY --from=build /work/falco-operator/LICENSE /usr/local/falco-operator/LICENSE

USER 65532:65532
ENTRYPOINT ["/usr/bin/manager"]

FROM ghcr.io/cybozu/ubuntu:24.04 AS falco-artifact-operator
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /usr/local/bin/falco-artifact-operator /usr/bin/manager
COPY --from=build /work/falco-operator/LICENSE /usr/local/falco-artifact-operator/LICENSE

USER 65532:65532

ENTRYPOINT ["/usr/bin/manager"]
