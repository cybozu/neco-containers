# syntax=docker/dockerfile:1
ARG GO_VERSION=1.26.7
ARG GO_AMD64_SHA256=ffb5f8de10c62550dfddab66b36b57030721e0a44a3218e9e1181d7b59f121ca
ARG GO_ARM64_SHA256=5a4ec883379d51ee9ce1040d5e87f8d35e20387574dd8c947feb01eabc3c1b37

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

FROM ghcr.io/cybozu/ubuntu-dev:22.04.20260422@sha256:03593cd228e0930638232fdf5a4afa23646320750286b6c08e9ad10292f7fcca
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
