# calico container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG CALICO_VERSION=3.8.2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p /go/src/github.com/projectcalico && \
    curl -sSLf https://github.com/projectcalico/node/archive/v${CALICO_VERSION}.tar.gz | \
      tar zxf - -C /go/src/github.com/projectcalico/ && \
    mv /go/src/github.com/projectcalico/node-${CALICO_VERSION} /go/src/github.com/projectcalico/node && \
    curl -sSLf https://github.com/projectcalico/typha/archive/v${CALICO_VERSION}.tar.gz | \
      tar zxf - -C /go/src/github.com/projectcalico/ && \
    mv /go/src/github.com/projectcalico/typha-${CALICO_VERSION} /go/src/github.com/projectcalico/typha

RUN mkdir -p /go/bin
RUN curl https://glide.sh/get | sh

WORKDIR /go/src/github.com/projectcalico/node
RUN glide install -strip-vendor
RUN GO111MODULE=on CGO_ENABLED=0 go install -mod=vendor ./cmd/calico-node

WORKDIR /go/src/github.com/projectcalico/typha
RUN glide install -strip-vendor
RUN GO111MODULE=on CGO_ENABLED=0 go install -mod=vendor ./cmd/calico-typha


# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

# Install remaining runtime deps required for felix from the global repository
RUN apt-get update && apt-get install -y --no-install-recommends \
    iptables \
    iproute2 \
    ipset \
    iputils-arping \
    iputils-ping \
    iputils-tracepath \
    # Need arp
    net-tools \
    conntrack \
    runit \
    # Need kmod to ensure ip6tables-save works correctly
    kmod \
    # Need netbase in order for ipset to work correctly
    # See https://github.com/kubernetes/kubernetes/issues/68703
    netbase \
    # Also needed (provides utilities for browsing procfs like ps)
    procps \
    && rm -rf /var/lib/apt/lists/*

# Copy in the filesystem - this contains felix, calico-bgp-daemon etc...
COPY --from=build /go/src/github.com/projectcalico/node/filesystem/ /

# Copy in the calico-node binary
COPY --from=build /go/bin/calico-node /usr/local/calico/bin/calico-node
COPY --from=build /go/bin/calico-typha /usr/local/calico/bin/calico-typha
COPY --from=build /go/src/github.com/projectcalico/node/LICENSE /usr/local/calico/LICENSE

# For typha
COPY --from=build /go/src/github.com/projectcalico/typha/docker-image/typha.cfg /etc/calico/typha.cfg

# Since our binary isn't designed to run as PID 1, run it via the tini init daemon.
ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-static-amd64 /usr/local/calico/bin/tini
RUN chmod +x /usr/local/calico/bin/tini

ENV PATH=/usr/local/calico/bin:"$PATH"
USER 10000:10000
