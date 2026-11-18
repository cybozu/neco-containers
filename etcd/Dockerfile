# etcd container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.11.2_resolute@sha256:afc977f11e560a1997622db9ee7974a5788019ea56a19f6600b7ac7c52e28794 AS build
ENV GOPROXY=https://golang.flatt.tech

# https://github.com/etcd-io/etcd/releases/tag/v3.6.17
ARG ETCD_COMMIT=0689586e360e2c6b46f95ad5634c25d5fa1683db

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/etcd-io/etcd/archive/${ETCD_COMMIT}.tar.gz | \
        tar zxf - \
    && mv etcd-${ETCD_COMMIT} etcd

WORKDIR /work/etcd
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    ./scripts/build.sh

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:26.04.20261117@sha256:47ff7a519b04321a84301061371358e06e97f80f4ef0229e5e2ad2e7f71acb55
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/etcd/bin /usr/local/etcd/bin
COPY --from=build /work/etcd/LICENSE /usr/local/etcd/LICENSE
COPY install-tools /usr/local/etcd/install-tools

ENV PATH=/usr/local/etcd/bin:"$PATH"

EXPOSE 2379 2380

RUN mkdir -p /var/lib/etcd && chown 10000:10000 /var/lib/etcd
USER 10000:10000
VOLUME /var/lib/etcd

ENTRYPOINT ["/usr/local/etcd/bin/etcd", "--data-dir", "/var/lib/etcd"]
