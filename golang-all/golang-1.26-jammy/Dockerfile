# syntax=docker/dockerfile:1

FROM scratch AS go-dist-amd64
ADD --checksum=sha256:1153d3d50e0ac764b447adfe05c2bcf08e889d42a02e0fe0259bd47f6733ad7f \
    https://dl.google.com/go/go1.26.4.linux-amd64.tar.gz /go.tar.gz

FROM scratch AS go-dist-arm64
ADD --checksum=sha256:ef758ae7c6cf9267c9c0ef080b8965f453d89ab2d25d9eb22de4405925238768 \
    https://dl.google.com/go/go1.26.4.linux-arm64.tar.gz /go.tar.gz

FROM go-dist-${TARGETARCH} AS go-dist

FROM ghcr.io/cybozu/ubuntu-dev:22.04.20260422@sha256:03593cd228e0930638232fdf5a4afa23646320750286b6c08e9ad10292f7fcca AS go-base
ARG TARGETARCH
ENV GOARCH=${TARGETARCH}
ENV GOTOOLCHAIN=local
ENV GOPATH=/go
ENV PATH=/go/bin:/usr/local/go/bin:$PATH

RUN --mount=type=bind,from=go-dist,source=/go.tar.gz,target=/tmp/go.tar.gz \
    rm -rf /usr/local/go \
    && tar -x -z -C /usr/local -f /tmp/go.tar.gz \
    && mkdir -p "$GOPATH/src" "$GOPATH/bin"

FROM go-base AS tools-builder
RUN --mount=type=bind,source=tools,target=/tmp/golang-tools \
    cd /tmp/golang-tools \
    && GOBIN=/go-tools go install tool

FROM go-base
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"
COPY --from=tools-builder /go-tools/ /usr/local/bin/

WORKDIR /work
CMD ["/bin/bash"]
