# etcd container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.16-focal AS build

ARG ETCD_VERSION=3.3.25

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p /go/src/github.com/coreos/ && \
    curl -sSLf https://github.com/coreos/etcd/archive/v${ETCD_VERSION}.tar.gz | \
      tar zxf - -C /go/src/github.com/coreos/ && \
    mv /go/src/github.com/coreos/etcd-${ETCD_VERSION} /go/src/github.com/coreos/etcd

WORKDIR /go/src/github.com/coreos/etcd
RUN GO111MODULE=off ./build

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /go/src/github.com/coreos/etcd/bin /usr/local/etcd/bin
COPY --from=build /go/src/github.com/coreos/etcd/LICENSE /usr/local/etcd/LICENSE
COPY install-tools /usr/local/etcd/install-tools

ENV PATH=/usr/local/etcd/bin:"$PATH"
ENV ETCDCTL_API=3

EXPOSE 2379 2380

RUN mkdir -p /var/lib/etcd && chown 10000:10000 /var/lib/etcd
USER 10000:10000
VOLUME /var/lib/etcd

ENTRYPOINT ["/usr/local/etcd/bin/etcd", "--data-dir", "/var/lib/etcd"]
