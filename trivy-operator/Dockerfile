# syntax=docker/dockerfile:1
# https://github.com/aquasecurity/trivy-operator/blob/main/build/trivy-operator/Dockerfile
# The upstream Dockerfile only copies a pre-built binary, so we build from source here.
ARG TRIVY_OPERATOR_VERSION=0.31.1
ARG TRIVY_OPERATOR_SHA256=5def17625660d43f843400fded79682b5feb172f3e5bd40f2d944a41c4b318f5

FROM scratch AS trivy-operator-src
ARG TRIVY_OPERATOR_VERSION
ARG TRIVY_OPERATOR_SHA256
ADD --checksum=sha256:${TRIVY_OPERATOR_SHA256} \
    https://github.com/aquasecurity/trivy-operator/archive/refs/tags/v${TRIVY_OPERATOR_VERSION}.tar.gz /src.tar.gz

FROM ghcr.io/cybozu/golang:1.26.4.1_noble@sha256:add9d704d4b75df2c51328615be89b61a3e71e4833321aa02c3f325a30d3eb8f AS builder
ARG TRIVY_OPERATOR_VERSION

WORKDIR /work/trivy-operator

RUN --mount=type=bind,from=trivy-operator-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .

RUN go mod download

ENV CGO_ENABLED=0
ENV GOEXPERIMENT=jsonv2
RUN go build -ldflags="-X main.version=${TRIVY_OPERATOR_VERSION}" \
    -o ./bin/trivy-operator ./cmd/trivy-operator/main.go

FROM gcr.io/distroless/static-debian12@sha256:9c346e4be81b5ca7ff31a0d89eaeade58b0f95cfd3baed1f36083ddb47ca3160
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/trivy-operator/bin/trivy-operator /usr/local/bin/
COPY --from=builder /work/trivy-operator/LICENSE /LICENSE
COPY --from=builder /work/trivy-operator/NOTICE /NOTICE

USER 10000:10000

ENTRYPOINT ["/usr/local/bin/trivy-operator"]
