# etcd container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.17-focal AS build

ARG ETCD_VERSION=3.5.3

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/etcd-io/etcd/archive/v${ETCD_VERSION}.tar.gz | \
        tar zxf - \
    && mv etcd-${ETCD_VERSION} etcd

WORKDIR /work/etcd
RUN ./build.sh

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /work/etcd/bin /usr/local/etcd/bin
COPY --from=build /work/etcd/LICENSE /usr/local/etcd/LICENSE
COPY install-tools /usr/local/etcd/install-tools

ENV PATH=/usr/local/etcd/bin:"$PATH"

EXPOSE 2379 2380

RUN mkdir -p /var/lib/etcd && chown 10000:10000 /var/lib/etcd
USER 10000:10000
VOLUME /var/lib/etcd

ENTRYPOINT ["/usr/local/etcd/bin/etcd", "--data-dir", "/var/lib/etcd"]
