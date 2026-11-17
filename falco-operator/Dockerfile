FROM ghcr.io/cybozu/golang:1.27-resolute AS build
ENV GOPROXY=https://golang.flatt.tech

ARG FALCO_OPERATOR_COMMIT=eda94565d62986674bc63499dc313df5c9a16f82 # v0.5.0
ARG IMAGE_REGISTRY=ghcr.io
ARG IMAGE_NAMESPACE=cybozu
ARG ARTIFACT_OPERATOR_TAG=0.5.0.1
ARG PROJECT=github.com/falcosecurity/falco-operator

WORKDIR /work/falco-operator
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    curl -fsSL -o /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz https://github.com/falcosecurity/falco-operator/archive/${FALCO_OPERATOR_COMMIT}.tar.gz && \
    tar -xzf /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz --strip-components=1 && \
    rm /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz && \
    CGO_ENABLED=0 go build -ldflags \
        "-s -w \
        -X '${PROJECT}/internal/pkg/image.Registry=${IMAGE_REGISTRY}' \
        -X '${PROJECT}/internal/pkg/image.Namespace=${IMAGE_NAMESPACE}' \
        -X '${PROJECT}/internal/pkg/version.ArtifactOperatorTag=${ARTIFACT_OPERATOR_TAG}'" \
        -o /usr/local/bin/falco-operator ./cmd/instance/main.go && \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o /usr/local/bin/artifact-operator ./cmd/artifact/main.go

FROM scratch AS falco-operator
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/local/bin/falco-operator /usr/bin/manager
COPY --from=build /work/falco-operator/LICENSE /usr/local/falco-operator/LICENSE

USER 65532:65532
ENTRYPOINT ["/usr/bin/manager"]

FROM scratch AS artifact-operator
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /usr/local/bin/artifact-operator /usr/bin/manager
COPY --from=build /work/falco-operator/LICENSE /usr/local/artifact-operator/LICENSE

USER 65532:65532
ENTRYPOINT ["/usr/bin/manager"]
