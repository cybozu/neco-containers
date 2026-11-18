FROM ghcr.io/cybozu/golang:1.26.11.2_resolute@sha256:afc977f11e560a1997622db9ee7974a5788019ea56a19f6600b7ac7c52e28794 AS build
ENV GOPROXY=https://golang.flatt.tech

# https://github.com/coredns/coredns/releases/tag/v1.14.7
ARG COREDNS_COMMIT=427fc80ed9ca47f354585eb30a3f1332950856c4

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/coredns/coredns/archive/${COREDNS_COMMIT}.tar.gz | \
        tar zxf - -C /work/ \
    && mkdir -p /go/src/github.com/coredns/ \
    && mv coredns-${COREDNS_COMMIT} /go/src/github.com/coredns/coredns

WORKDIR /go/src/github.com/coredns/coredns/
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    make

FROM ghcr.io/cybozu/ubuntu:26.04.20261117@sha256:47ff7a519b04321a84301061371358e06e97f80f4ef0229e5e2ad2e7f71acb55
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/src/github.com/coredns/coredns/LICENSE /usr/local/coredns/LICENSE
COPY --from=build /go/src/github.com/coredns/coredns/coredns /usr/local/coredns/bin/coredns
ENV PATH=/usr/local/coredns/bin:"$PATH"

USER 10000:10000
WORKDIR /
EXPOSE 1053 1053/udp
ENTRYPOINT ["coredns"]
