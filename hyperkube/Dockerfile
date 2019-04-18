# hyperkube container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG K8S_VERSION=1.14.1

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
    rsync \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/k8s.io/ && \
    curl -sSLf https://github.com/kubernetes/kubernetes/archive/v${K8S_VERSION}.tar.gz | \
      tar zxf - -C /go/src/k8s.io/ && \
    mv /go/src/k8s.io/kubernetes-${K8S_VERSION} /go/src/k8s.io/kubernetes

WORKDIR /go/src/k8s.io/kubernetes
RUN make WHAT=cmd/hyperkube GOLDFLAGS="-w -s"

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
    btrfs-progs \
    ceph-common \
    conntrack \
    ebtables \
    git \
    iproute2 \
    ipset \
    iptables \
    ipvsadm \
    kmod \
    openssh-client \
    socat \
    udev \
    xfsprogs \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /go/src/k8s.io/kubernetes/_output/bin/hyperkube /usr/local/kubernetes/bin/hyperkube
COPY --from=build /go/src/k8s.io/kubernetes/LICENSE /usr/local/kubernetes/LICENSE
COPY install-tools /usr/local/kubernetes/install-tools

RUN ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/kube-apiserver \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/apiserver \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/kube-controller-manager \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/controller-manager \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/kubectl \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/kubelet \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/kube-proxy \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/proxy \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/kube-scheduler \
 && ln -s /usr/local/kubernetes/bin/hyperkube /usr/local/kubernetes/bin/scheduler

ENV PATH=/usr/local/kubernetes/bin:"$PATH"
