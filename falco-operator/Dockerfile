FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG FALCO_OPERATOR_COMMIT=17dc659e26bc743b49aff099d585b02ec6809a5b # v0.2.1

WORKDIR /work/falco-operator
RUN curl -fsSL -o /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz https://github.com/falcosecurity/falco-operator/archive/${FALCO_OPERATOR_COMMIT}.tar.gz && \
    tar -xzf /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz --strip-components=1 && \
    rm /tmp/falco-operator-${FALCO_OPERATOR_COMMIT}.tar.gz && \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o /usr/local/bin/falco-operator ./cmd/instance/main.go && \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o /usr/local/bin/falco-artifact-operator ./cmd/artifact/main.go

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /usr/local/bin/falco-operator /usr/bin/falco-operator
COPY --from=build /usr/local/bin/falco-artifact-operator /usr/bin/falco-artifact-operator
COPY --from=build /work/falco-operator/LICENSE /usr/local/falco-operator/LICENSE

USER 65532:65532
