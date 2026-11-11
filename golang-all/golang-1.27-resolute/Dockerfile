# syntax=docker/dockerfile:1
ARG GO_VERSION=1.27.4
ARG GO_AMD64_SHA256=7226736c2339719cda28fbf3a54d71aa63bcefe376e98cef00034268dab24840
ARG GO_ARM64_SHA256=3392582988a273a87ab5043ee1b2c9c17a2c3c00580bf9326c47a2e0aaaeac92

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

FROM ghcr.io/cybozu/ubuntu-dev:26.04.20261021@sha256:e8cea4ac6112e219658cda415a875c72d5ae9ed8a6c7df4a070e7777a558bb80
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
