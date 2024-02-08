ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.20-jammy
ARG UBUNTU_IMAGE=ghcr.io/cybozu/ubuntu:22.04
ARG DESTDIR=/tmp/install/linux/amd64

# Stage1: build common
FROM ${GOLANG_IMAGE} as build-base
ARG DESTDIR
ENV DESTDIR=${DESTDIR}
COPY TAG /

WORKDIR /go/src/github.com/cybozu-go/
RUN mkdir -p ${DESTDIR} \
    && VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    # Since we use the fork and patched repogitory, we need to specify the branch name
    && BRANCH=v${VERSION}-lb-dsr-patch \
    && git clone --depth 1 --branch ${BRANCH} https://github.com/cybozu-go/cilium \
    && cd cilium \
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
    && cp images/runtime/iptables-wrapper-installer.sh \
      images/cilium/init-container.sh \
      plugins/cilium-cni/cni-install.sh \
      plugins/cilium-cni/install-plugin.sh \
      plugins/cilium-cni/cni-uninstall.sh \
        ${DESTDIR} \
    && images/cilium/download-hubble.sh


FROM build-base as builder
COPY src/workspace/bin/llvm-objcopy /bin/
COPY src/workspace/bin/clang src/workspace/bin/llc /bin/
WORKDIR /go/src/github.com/cybozu-go/cilium
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
      libz3-dev \
    && images/builder/install-protoc.sh \
    && make build-container install-container-binary

FROM ${UBUNTU_IMAGE}
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG DESTDIR
COPY src/workspace/bin/clang src/workspace/bin/llc /bin/
COPY src/workspace/usr/local/bin /usr/local/bin
COPY src/workspace/usr/bin /usr/bin
COPY src/workspace/usr/lib /usr/lib
# When used within the Cilium container, Hubble CLI should target the
# local unix domain socket instead of Hubble Relay.
ENV HUBBLE_SERVER=unix:///var/run/cilium/hubble.sock
# This version of Cilium originally shipped with an older verison of the
# Hubble CLI using legacy JSON output which we want to preserve.
ENV HUBBLE_COMPAT=legacy-json-output
COPY --from=build-base /out/linux/amd64/bin/loopback /cni/loopback
COPY --from=build-base /out/linux/amd64/bin/gops /bin/gops
COPY --from=build-base /out/linux/amd64/bin/hubble /usr/bin/hubble
COPY --from=builder ${DESTDIR} /

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      libelf1 \
      libmnl0 \
      bash-completion \
      iptables \
      ipset \
      kmod \
      ca-certificates \
      jq \
    && apt-get purge --auto-remove \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && /iptables-wrapper-installer.sh --no-sanity-check

WORKDIR /home/cilium

ENV INITSYSTEM="SYSTEMD"
ENTRYPOINT ["/usr/bin/cilium"]
