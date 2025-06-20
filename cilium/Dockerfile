# Image Name Mapping:
#   cilium-compiler:            docker.io/cilium/image-compilers
#   cilium-llvm:                docker.io/cilium/cilium-llvm
#   cilium-bpftool              docker.io/cilium/cilium-bpftool
#   cilium-iptables             intermediate (https://github.com/cilium/image-tools/blob/master/images/iptables/Dockerfile)
#   gops-cni-builder            intermediate (https://github.com/cilium/cilium/blob/main/images/runtime/Dockerfile)
#   cilium-runtime              quay.io/cilium/cilium-runtime
#   cilium-builder              quay.io/cilium/cilium-builder
#   cilium-hubble               intermediate (https://github.com/cilium/cilium/blob/main/images/cilium/Dockerfile)
#   cilium-agent-builder        intermediate (https://github.com/cilium/cilium/blob/main/images/cilium/Dockerfile)
#
# ref: https://docs.cilium.io/en/stable/contributing/development/images/

ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.24-noble
ARG UBUNTU_IMAGE=ghcr.io/cybozu/ubuntu:24.04
ARG DESTDIR=/tmp/install/linux/amd64

# -----------------------------------
FROM ${GOLANG_IMAGE} AS cybozu-golang

# ------------------------------------------------------------------------------------------
# https://github.com/cilium/image-tools/blob/master/images/compilers/Dockerfile (AS builder)
FROM ${UBUNTU_IMAGE} AS cilium-compiler
COPY src/image-tools/images/compilers/install-deps.sh /tmp/install-deps.sh

# Workaround
RUN mv /tmp/install-deps.sh /tmp/install-deps-original.sh \
    && cat /tmp/install-deps-original.sh | grep -v libelf-dev:arm64 > /tmp/install-deps.sh \
    && chmod +x /tmp/install-deps.sh

RUN /tmp/install-deps.sh

# -------------------------------------------------------------------------------------
# https://github.com/cilium/image-tools/blob/master/images/llvm/Dockerfile (AS builder)
FROM cilium-compiler AS cilium-llvm
COPY src/image-tools/images/llvm/checkout-llvm.sh /tmp/checkout-llvm.sh
COPY src/image-tools/images/llvm/build-llvm-native.sh /tmp/build-llvm-native.sh

RUN /tmp/checkout-llvm.sh
RUN /tmp/build-llvm-native.sh
RUN cp /out/linux/amd64/bin/* /usr/local/bin/

# ----------------------------------------------------------------------------------------
# https://github.com/cilium/image-tools/blob/master/images/bpftool/Dockerfile (AS builder)
FROM cilium-compiler AS cilium-bpftool
COPY src/image-tools/images/bpftool/checkout-linux.sh /tmp/checkout-linux.sh
COPY src/image-tools/images/bpftool/build-bpftool-native.sh /tmp/build-bpftool-native.sh

RUN /tmp/checkout-linux.sh
RUN /tmp/build-bpftool-native.sh

RUN apt-get update \
    && apt-get install -y --no-install-recommends libelf1 \
    && apt-get purge --auto-remove -y

# -----------------------------------------------------------------------------------
# https://github.com/cilium/image-tools/blob/master/images/iptables/Dockerfile (AS .)
FROM ${UBUNTU_IMAGE} AS cilium-iptables
ENV IPTABLES_VERSION=1.8.8-1
ENV SNAPSHOT_DATE=20230116T212610Z

RUN mkdir /iptables
WORKDIR /iptables

RUN apt-get update && \
    apt-get install -y --no-install-recommends debian-archive-keyring apt-src ca-certificates && \
    echo "deb-src [check-valid-until=no signed-by=/usr/share/keyrings/debian-archive-bullseye-automatic.gpg] https://snapshot.debian.org/archive/debian/${SNAPSHOT_DATE}/ bookworm main" > /etc/apt/sources.list.d/iptables-snapshot.list && \
    apt-get update && \
    apt-src -b install iptables=${IPTABLES_VERSION} && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# ------------------------------------------------------------------------------------------
# https://github.com/cilium/cilium/blob/main/images/runtime/Dockerfile (AS gops-cni-builder)
FROM ${GOLANG_IMAGE} AS gops-cni-builder
RUN apt-get update && apt-get install -y binutils-aarch64-linux-gnu binutils-x86-64-linux-gnu

COPY src/cilium/images/runtime/build-gops.sh build-gops.sh
COPY src/cilium/images/runtime/cni-version.sh cni-version.sh
COPY src/cilium/images/runtime/download-cni.sh download-cni.sh

RUN ./build-gops.sh
RUN ./download-cni.sh

# --------------------------------------------------------------------------------
# https://github.com/cilium/cilium/blob/main/images/runtime/Dockerfile (AS rootfs)
# !! WARNING !!
# This is the base of the final image, do not leave anything unrelated to the final output.
FROM ${UBUNTU_IMAGE} AS cilium-runtime

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y jq && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/cilium/cilium/images/runtime
RUN --mount=type=bind,readwrite,source=src/cilium/images/runtime,target=/go/src/github.com/cilium/cilium/images/runtime \
    ./install-runtime-deps.sh

COPY --from=cilium-iptables /iptables /iptables
RUN dpkg -i /iptables/*\.deb && rm -rf /iptables

RUN --mount=type=bind,readwrite,source=src/cilium/images/runtime,target=/go/src/github.com/cilium/cilium/images/runtime \
    ./iptables-wrapper-installer.sh --no-sanity-check

COPY --from=cilium-llvm /usr/local/bin/clang /usr/local/bin/llc /usr/local/bin/
COPY --from=cilium-bpftool /out/linux/amd64/bin/bpftool /usr/local/bin/
COPY --from=gops-cni-builder /out/linux/amd64/bin/loopback /cni/loopback
COPY --from=gops-cni-builder /out/linux/amd64/bin/gops /bin/gops

# --------------------------------------------------------------------------------
# https://github.com/cilium/cilium/blob/main/images/builder/Dockerfile (AS rootfs)
FROM cilium-runtime AS cilium-builder

# TARGETARCH is an automatic platform ARG enabled by Docker BuildKit.
ARG TARGETARCH
RUN apt-get update && \
    apt-get upgrade -y --no-install-recommends && \
    apt-get install -y --no-install-recommends \
      # Install cross tools for both arm64 on amd64
      gcc-aarch64-linux-gnu \
      g++-aarch64-linux-gnu \
      libc6-dev-arm64-cross \
      binutils-aarch64-linux-gnu \
      gcc-x86-64-linux-gnu \
      g++-x86-64-linux-gnu \
      libc6-dev-amd64-cross \
      binutils-x86-64-linux-gnu \
      # Dependencies to unzip protoc
      unzip \
      # Base Cilium-build dependencies
      binutils \
      coreutils \
      curl \
      gcc \
      git \
      libc6-dev \
      make && \
    apt-get purge --auto-remove && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=cilium-compiler /usr/lib/aarch64-linux-gnu /usr/lib/aarch64-linux-gnu
COPY --from=cybozu-golang /usr/local/go /usr/local/go

RUN mkdir -p /go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH "${GOROOT}/bin:${GOPATH}/bin:${PATH}"

RUN CGO_ENABLED=0 go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /go/src/github.com/cilium/cilium/images/builder
RUN --mount=type=bind,readwrite,source=src/cilium/images/builder,target=/go/src/github.com/cilium/cilium/images/builder \
    ./install-gitconfig.sh

WORKDIR /go/src/github.com/cilium/cilium/images/builder
RUN --mount=type=bind,readwrite,source=src/cilium/images/builder,target=/go/src/github.com/cilium/cilium/images/builder \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg \
    ./build-debug-wrapper.sh

WORKDIR /go/src/github.com/cilium/cilium/images/builder
RUN --mount=type=bind,readwrite,source=src/cilium/images/builder,target=/go/src/github.com/cilium/cilium/images/builder \
    ./install-protoc.sh

RUN --mount=type=bind,readwrite,source=src/cilium/images/builder,target=/go/src/github.com/cilium/cilium/images/builder \
    ./install-protoplugins.sh

# used to facilitate the verifier tests
COPY --from=cilium-llvm /usr/local/bin/llvm-objcopy /bin/

# -------------------------------------------------------------------------------
# https://github.com/cilium/cilium/blob/main/images/cilium/Dockerfile (AS hubble)
FROM cilium-builder AS cilium-hubble
COPY src/cilium/images/cilium/download-hubble.sh /tmp/download-hubble.sh

RUN /tmp/download-hubble.sh
RUN /out/linux/amd64/bin/hubble completion bash > /out/linux/bash_completion

# --------------------------------------------------------------------------------
# https://github.com/cilium/cilium/blob/main/images/cilium/Dockerfile (AS builder)
FROM cilium-builder AS cilium-agent-builder
WORKDIR /go/src/github.com/cilium/cilium

RUN --mount=type=bind,readwrite,source=src/cilium,target=/go/src/github.com/cilium/cilium \
    make GOARCH=amd64 DESTDIR=/tmp/install/linux/amd64 PKG_BUILD=1 \
    build-container install-container-binary

RUN --mount=type=bind,readwrite,source=src/cilium,target=/go/src/github.com/cilium/cilium \
    make GOARCH=amd64 DESTDIR=/tmp/install/linux/amd64 PKG_BUILD=1 \
    install-bash-completion licenses-all && \
    mv LICENSE.all /tmp/install/linux/amd64/LICENSE.all

COPY src/cilium/images/cilium/init-container.sh \
     src/cilium/plugins/cilium-cni/install-plugin.sh \
     src/cilium/plugins/cilium-cni/cni-uninstall.sh \
       /tmp/install/linux/amd64

# --------------------------------------------------------------------------------
# https://github.com/cilium/cilium/blob/main/images/cilium/Dockerfile (AS release)
FROM cilium-runtime
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

## When used within the Cilium container, Hubble CLI should target the
## local unix domain socket instead of Hubble Relay.
ENV HUBBLE_SERVER=unix:///var/run/cilium/hubble.sock

RUN echo ". /etc/profile.d/bash_completion.sh" >> /etc/bash.bashrc
COPY --from=cilium-hubble /out/linux/amd64/bin/hubble /usr/bin/hubble
COPY --from=cilium-hubble /out/linux/bash_completion /etc/bash_completion.d/hubble
COPY --from=cilium-agent-builder /tmp/install/linux/amd64/ /

WORKDIR /home/cilium
ENV INITSYSTEM="SYSTEMD"
ENTRYPOINT ["/usr/bin/cilium-dbg"]
