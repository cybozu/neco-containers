ARG BASE_IMAGE=quay.io/cybozu/ubuntu:20.04
ARG GOLANG_IMAGE=quay.io/cybozu/golang:1.17-focal
ARG UBUNTU_IMAGE=quay.io/cybozu/ubuntu:20.04
ARG DESTDIR=/tmp/install/linux/amd64

# Stage1: build common
FROM ${GOLANG_IMAGE} as build-base
ARG DESTDIR
ENV DESTDIR=${DESTDIR}
COPY TAG /
COPY 17418.patch /tmp/

WORKDIR /go/src/github.com/cilium/cilium
RUN mkdir -p ${DESTDIR} \
    && VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL "https://github.com/cilium/cilium/archive/v${VERSION}.tar.gz" | \
      tar xzf - --strip-components 1 \
    && patch -p1 --no-backup-if-mismatch < /tmp/17418.patch \
    && make licenses-all \
    && mv LICENSE.all ${DESTDIR}/LICENSE \
    && apt-get update \
    && apt-get install -y --no-install-recommends binutils-aarch64-linux-gnu \
      libelf1 \
      libmnl0 \
      iptables \
      ipset \
      kmod \
    && images/runtime/build-gops.sh \
    && images/runtime/download-cni.sh \
    && mkdir -p ${DESTDIR}/usr/sbin \
    && cp images/runtime/iptables-wrapper ${DESTDIR}/usr/sbin/iptables-wrapper \
    && cp images/runtime/configure-iptables-wrapper.sh \
      images/cilium/init-container.sh \
      plugins/cilium-cni/cni-install.sh \
      plugins/cilium-cni/cni-uninstall.sh \
        ${DESTDIR} \
    && images/cilium/download-hubble.sh


FROM build-base as builder
COPY workspace/bin/llvm-objcopy /bin/
WORKDIR /go/src/github.com/cilium/cilium
ARG LIBNETWORK_PLUGIN
ARG DESTDIR
ENV PKG_BUILD=1
ENV SKIP_DOCS=true
ENV DESTDIR=${DESTDIR}
ENV LIBNETWORK_PLUGIN=${LIBNETWORK_PLUGIN}
RUN apt-get install -y --no-install-recommends binutils \
      binutils \
      coreutils \
      curl \
      gcc \
      git \
      libc6-dev \
      libelf-dev \
      make \
      unzip \
    && images/builder/install-protoc.sh \
    && make build-container install-container-binary

FROM ${UBUNTU_IMAGE}
ARG DESTDIR
COPY workspace/bin/clang workspace/bin/llc /bin/
COPY workspace/usr/local/bin /usr/local/bin
COPY workspace/usr/lib /usr/lib
# When used within the Cilium container, Hubble CLI should target the
# local unix domain socket instead of Hubble Relay.
ENV HUBBLE_SERVER=unix:///var/run/cilium/hubble.sock
COPY --from=build-base /out/linux/amd64/bin/loopback /cni/loopback
COPY --from=build-base /out/linux/amd64/bin/gops /bin/gops
COPY --from=build-base /out/linux/amd64/bin/hubble /usr/bin/hubble
COPY --from=build-base ${DESTDIR} /

RUN apt-get update \
    && apt-get install -y --no-install-recommends libelf1 \
      libmnl0 \
      iptables \
      ipset \
      kmod \
    && apt-get purge --auto-remove \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && /configure-iptables-wrapper.sh

WORKDIR /home/cilium

ENV INITSYSTEM="SYSTEMD"
ENTRYPOINT ["/cilium"]
