# kubernetes container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.11.2_resolute@sha256:afc977f11e560a1997622db9ee7974a5788019ea56a19f6600b7ac7c52e28794 AS build

# https://github.com/kubernetes/kubernetes/releases/tag/v1.36.6
ARG K8S_COMMIT=1b2171b110bb2a2542e1322035c6583ec86eb976

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update \
    && apt-get -y install --no-install-recommends \
    rsync \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/k8s.io/ && \
    curl -sSLf https://github.com/kubernetes/kubernetes/archive/${K8S_COMMIT}.tar.gz | \
    tar zxf - -C /go/src/k8s.io/ && \
    mv /go/src/k8s.io/kubernetes-${K8S_COMMIT} /go/src/k8s.io/kubernetes

COPY kubelet-fetch-active-container.diff /work/

WORKDIR /go/src/k8s.io/kubernetes
RUN patch -p1 < /work/kubelet-fetch-active-container.diff
RUN make all WHAT="cmd/kube-apiserver cmd/kube-controller-manager cmd/kube-proxy cmd/kube-scheduler cmd/kubelet" GOLDFLAGS="-w -s"

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:26.04.20261117@sha256:47ff7a519b04321a84301061371358e06e97f80f4ef0229e5e2ad2e7f71acb55
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
