# calico container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG CALICO_VERSION=3.7.2

RUN mkdir -p /go/src/github.com/projectcalico && \
    curl -sSLf https://github.com/projectcalico/felix/archive/v${CALICO_VERSION}.tar.gz | \
      tar zxf - -C /go/src/github.com/projectcalico/ && \
    mv /go/src/github.com/projectcalico/felix-${CALICO_VERSION} /go/src/github.com/projectcalico/felix && \
    curl -sSLf https://github.com/projectcalico/typha/archive/v${CALICO_VERSION}.tar.gz | \
      tar zxf - -C /go/src/github.com/projectcalico/ && \
    mv /go/src/github.com/projectcalico/typha-${CALICO_VERSION} /go/src/github.com/projectcalico/typha

WORKDIR /go/src/github.com/projectcalico/felix
RUN GO111MODULE=on CGO_ENABLED=0 go install ./cmd/calico-felix

WORKDIR /go/src/github.com/projectcalico/typha
RUN GO111MODULE=on CGO_ENABLED=0 go install ./cmd/calico-typha

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

# Copy in the calico-node binary
COPY --from=build /go/bin/calico-felix /bin/calico-felix
COPY --from=build /go/bin/calico-typha /bin/calico-typha

# For typha
COPY --from=build /go/src/github.com/projectcalico/typha/docker-image/typha.cfg /etc/calico/typha.cfg

# For felix
COPY --from=build /go/src/github.com/projectcalico/felix/docker-image/calico-felix-wrapper /usr/bin/calico-felix-wrapper

# Since our binary isn't designed to run as PID 1, run it via the tini init daemon.
ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-static-amd64 /sbin/tini
RUN chmod +x /sbin/tini
ENTRYPOINT ["/sbin/tini", "--"]

#CMD ["/usr/bin/calico-felix-wrapper"]
#CMD ["/usr/bin/calico-typha"]
