# syntax=docker/dockerfile:1
ARG GO_VERSION=1.26.6
ARG GO_AMD64_SHA256=708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89
ARG GO_ARM64_SHA256=d0507e9e9d7fe012aae570108cbd76c15de879e17130ab8cb90d4d7445cb1f2e

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
