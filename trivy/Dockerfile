# syntax=docker/dockerfile:1
# https://github.com/aquasecurity/trivy/blob/main/Dockerfile
# The upstream Dockerfile only copies a pre-built binary, so we build from source here.
ARG TRIVY_VERSION=0.71.1
# TODO: Fill this with the sha256 of:
# https://github.com/aquasecurity/trivy/archive/refs/tags/v${TRIVY_VERSION}.tar.gz
ARG TRIVY_SHA256=fb79664621120e700b89f6a4642204787d43cf24d7cec8f37d8a39d26e8bfa9d

FROM scratch AS trivy-src
ARG TRIVY_VERSION
ARG TRIVY_SHA256
ADD --checksum=sha256:${TRIVY_SHA256} \
    https://github.com/aquasecurity/trivy/archive/refs/tags/v${TRIVY_VERSION}.tar.gz /src.tar.gz

FROM ghcr.io/cybozu/golang:1.26.4.1_noble@sha256:add9d704d4b75df2c51328615be89b61a3e71e4833321aa02c3f325a30d3eb8f AS builder
ARG TRIVY_VERSION

WORKDIR /work/trivy

RUN --mount=type=bind,from=trivy-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .

RUN go mod download

ENV CGO_ENABLED=0
ENV GOEXPERIMENT=jsonv2

RUN go build \
    -ldflags="-s -w -extldflags '-static' -X github.com/aquasecurity/trivy/pkg/version/app.ver=${TRIVY_VERSION}" \
    -o ./bin/trivy ./cmd/trivy/

FROM gcr.io/distroless/static-debian12@sha256:9c346e4be81b5ca7ff31a0d89eaeade58b0f95cfd3baed1f36083ddb47ca3160
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/trivy/bin/trivy /usr/local/bin/trivy
COPY --from=builder /work/trivy/LICENSE /LICENSE
COPY --from=builder /work/trivy/contrib/*.tpl /contrib/

# Trivy downloads vulnerability DB/cache at runtime.
# With a non-root user, use a writable cache directory.
ENV TRIVY_CACHE_DIR=/tmp/trivy

USER 10000:10000

ENTRYPOINT ["/usr/local/bin/trivy"]
