# etcd container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

# v3.6.11
ARG ETCD_COMMIT=ec166e2292a58365c90e96fcd206b3b74938d49d

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/etcd-io/etcd/archive/${ETCD_COMMIT}.tar.gz | \
        tar zxf - \
    && mv etcd-${ETCD_COMMIT} etcd

WORKDIR /work/etcd
RUN ./scripts/build.sh

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04
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
