# containerd-installer container
# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG CONTAINERD_VERSION=1.2.4
ARG PROTOBUF_VERSION=3.6.1
ARG CRITOOLS_VERSION=1.13.0

WORKDIR /work

# Build containerd
RUN apt-get update \
&& apt-get -y install --no-install-recommends libseccomp-dev unzip \
&& rm -rf /var/lib/apt/lists/*

RUN curl -fsSL -o protoc.zip https://github.com/google/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip \
&& unzip -o protoc.zip -d /usr/local \
&& rm -f protoc.zip

RUN curl -fsSL -o containerd.tar.gz https://github.com/containerd/containerd/archive/v${CONTAINERD_VERSION}.tar.gz \
&& mkdir -p ${GOPATH}/src/github.com/containerd/containerd \
&& tar -x -z --strip-components 1 -f containerd.tar.gz -C ${GOPATH}/src/github.com/containerd/containerd \
&& rm -f containerd.tar.gz \
&& cd ${GOPATH}/src/github.com/containerd/containerd \
&& make VERSION=${CONTAINERD_VERSION} BUILDTAGS=no_btrfs

# Build cri-tools
RUN curl -fsSL -o crictl.tar.gz https://github.com/kubernetes-sigs/cri-tools/archive/v${CRITOOLS_VERSION}.tar.gz \
&& mkdir -p ${GOPATH}/src/github.com/kubernetes-sigs/cri-tools \
&& tar -x -z --strip-components 1 -f crictl.tar.gz -C ${GOPATH}/src/github.com/kubernetes-sigs/cri-tools \
&& cd ${GOPATH}/src/github.com/kubernetes-sigs/cri-tools \
&& make VERSION=${CRITOOLS_VERSION} crictl

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/src/github.com/containerd/containerd/bin /usr/local/containerd-installer/bin
COPY --from=build /go/src/github.com/containerd/containerd/LICENSE /usr/local/containerd-installer/LICENSE-containerd
COPY --from=build /go/bin/crictl /usr/local/containerd-installer/bin/crictl
COPY --from=build /go/src/github.com/kubernetes-sigs/cri-tools/LICENSE /usr/local/containerd-installer/LICENSE-cri-tools
COPY install-tools /usr/local/containerd-installer/install-tools
