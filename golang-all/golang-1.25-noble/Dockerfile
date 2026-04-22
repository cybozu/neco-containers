# syntax=docker/dockerfile:1

FROM scratch AS go-dist-amd64
ADD --checksum=sha256:00859d7bd6defe8bf84d9db9e57b9a4467b2887c18cd93ae7460e713db774bc1 \
    https://dl.google.com/go/go1.25.9.linux-amd64.tar.gz /go.tar.gz

FROM scratch AS go-dist-arm64
ADD --checksum=sha256:ec342e7389b7f489564ed5463c63b16cf8040023dabc7861256677165a8c0e2b \
    https://dl.google.com/go/go1.25.9.linux-arm64.tar.gz /go.tar.gz

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
