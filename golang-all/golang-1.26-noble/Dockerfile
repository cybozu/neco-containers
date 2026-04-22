# syntax=docker/dockerfile:1

FROM scratch AS go-dist-amd64
ADD --checksum=sha256:990e6b4bbba816dc3ee129eaeaf4b42f17c2800b88a2166c265ac1a200262282 \
    https://dl.google.com/go/go1.26.2.linux-amd64.tar.gz /go.tar.gz

FROM scratch AS go-dist-arm64
ADD --checksum=sha256:c958a1fe1b361391db163a485e21f5f228142d6f8b584f6bef89b26f66dc5b23 \
    https://dl.google.com/go/go1.26.2.linux-arm64.tar.gz /go.tar.gz

FROM go-dist-${TARGETARCH} AS go-dist

FROM ghcr.io/cybozu/ubuntu-dev:24.04.20260422@sha256:d9dcc6f47039bbd1faeb6dcdc75f3e469756e481abbf8535201783fc68ff611d AS go-base
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
