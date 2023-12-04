# kubernetes container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.21-jammy AS build

ARG K8S_VERSION=1.27.8

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update \
    && apt-get -y install --no-install-recommends \
    rsync \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/k8s.io/ && \
    curl -sSLf https://github.com/kubernetes/kubernetes/archive/v${K8S_VERSION}.tar.gz | \
    tar zxf - -C /go/src/k8s.io/ && \
    mv /go/src/k8s.io/kubernetes-${K8S_VERSION} /go/src/k8s.io/kubernetes

WORKDIR /go/src/k8s.io/kubernetes
RUN make all WHAT="cmd/kube-apiserver cmd/kube-controller-manager cmd/kube-proxy cmd/kube-scheduler cmd/kubelet" GOLDFLAGS="-w -s"

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
    btrfs-progs \
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

COPY --from=build /go/src/k8s.io/kubernetes/_output/bin/kube-apiserver /usr/local/kubernetes/bin/kube-apiserver
COPY --from=build /go/src/k8s.io/kubernetes/_output/bin/kube-controller-manager /usr/local/kubernetes/bin/kube-controller-manager
COPY --from=build /go/src/k8s.io/kubernetes/_output/bin/kube-proxy /usr/local/kubernetes/bin/kube-proxy
COPY --from=build /go/src/k8s.io/kubernetes/_output/bin/kube-scheduler /usr/local/kubernetes/bin/kube-scheduler
COPY --from=build /go/src/k8s.io/kubernetes/_output/bin/kubelet /usr/local/kubernetes/bin/kubelet
COPY --from=build /go/src/k8s.io/kubernetes/LICENSE /usr/local/kubernetes/LICENSE

ENV PATH=/usr/local/kubernetes/bin:"$PATH"
