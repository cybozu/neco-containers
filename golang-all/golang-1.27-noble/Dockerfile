# syntax=docker/dockerfile:1
ARG GO_VERSION=1.27.2
ARG GO_AMD64_SHA256=ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5
ARG GO_ARM64_SHA256=94f3e30b8e374bc285e7dadc11e0865726b9bc6e85b841ccceaabc0214c6b7c8

FROM scratch AS go-dist-amd64
ARG GO_VERSION
ARG GO_AMD64_SHA256
ADD --checksum=sha256:${GO_AMD64_SHA256} \
    https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz /go.tar.gz

FROM scratch AS go-dist-arm64
ARG GO_VERSION
ARG GO_ARM64_SHA256
ADD --checksum=sha256:${GO_ARM64_SHA256} \
    https://dl.google.com/go/go${GO_VERSION}.linux-arm64.tar.gz /go.tar.gz

FROM go-dist-${TARGETARCH} AS go-dist

FROM ghcr.io/cybozu/ubuntu-dev:24.04.20260902@sha256:d1d788d7de6d89d8253484b4f6d959a86005aee02ef2b69d9f451fba86be6a4d
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"
ARG TARGETARCH
ENV GOARCH=${TARGETARCH}
ENV GOTOOLCHAIN=local
ENV GOPATH=/go
ENV PATH=/go/bin:/usr/local/go/bin:$PATH

RUN --mount=type=bind,from=go-dist,source=/go.tar.gz,target=/tmp/go.tar.gz \
    rm -rf /usr/local/go \
    && tar -x -z -C /usr/local -f /tmp/go.tar.gz \
    && mkdir -p "$GOPATH/src" "$GOPATH/bin"

WORKDIR /work
CMD ["/bin/bash"]
