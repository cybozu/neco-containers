# Image Name Mapping:
#   cilium-compiler:            docker.io/cilium/image-compilers
#   cilium-llvm:                docker.io/cilium/cilium-llvm
#   cilium-bpftool              docker.io/cilium/cilium-bpftool
#   cilium-iptables             intermediate (https://github.com/cilium/image-tools/blob/master/images/iptables/Dockerfile)
#   gops-cni-builder            intermediate (https://github.com/cilium/cilium/blob/main/images/runtime/Dockerfile)
#   cilium-runtime              quay.io/cilium/cilium-runtime
#   cilium-builder              quay.io/cilium/cilium-builder
#   cilium-agent-builder        intermediate (https://github.com/cilium/cilium/blob/main/images/cilium/Dockerfile)
#
# ref: https://docs.cilium.io/en/stable/contributing/development/images/

ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.26-noble
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

ARG LLVM_VERSION="llvmorg-19.1.7"
ADD https://github.com/llvm/llvm-project/archive/${LLVM_VERSION}.tar.gz /tmp/llvm.tar.gz
WORKDIR /src/llvm
RUN tar -xf /tmp/llvm.tar.gz --strip-components=1 --directory /src/llvm

WORKDIR /src/llvm/llvm/build
COPY src/image-tools/images/llvm/build-llvm.sh /tmp/build-llvm.sh
RUN /tmp/build-llvm.sh

RUN cp /out/bin/* /usr/local/bin/

# ----------------------------------------------------------------------------------------
# https://github.com/cilium/image-tools/blob/master/images/bpftool/Dockerfile (AS builder)
FROM cilium-compiler AS cilium-bpftool

ARG BPFTOOL_VERSION="v7.4.0"
WORKDIR /bpftool
RUN git clone --recurse-submodules --branch "${BPFTOOL_VERSION}" https://github.com/libbpf/bpftool.git /bpftool

RUN \
    # From Ubuntu 24.04 builder image, libzstd must be added at the end of LIBS and LIBS_BOOTSTRAP to compile statically
    # See https://github.com/libbpf/bpftool/issues/152
    sed -i "s/\(LIBS = \$(LIBBPF) -lelf -lz\)/\1 -lzstd/; s/\(LIBS_BOOTSTRAP = \$(LIBBPF_BOOTSTRAP) -lelf -lz\)/\1 -lzstd/" /bpftool/src/Makefile \
    && make -C src EXTRA_CFLAGS=--static BPFTOOL_VERSION="${BPFTOOL_VERSION#v}" -j "$(nproc)" \
    && strip /bpftool/src/bpftool \
    && mkdir -p /out/bin \
    && mv /bpftool/src/bpftool /out/bin/

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
    apt-src -b install iptables="${IPTABLES_VERSION}" && \
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
COPY --from=cilium-bpftool /out/bin/bpftool /usr/local/bin/
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
      make \
      ccache && \
    apt-get purge --auto-remove && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=cilium-compiler /usr/lib/aarch64-linux-gnu /usr/lib/aarch64-linux-gnu
COPY --from=cybozu-golang /usr/local/go /usr/local/go

RUN mkdir -p /go
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH="${GOROOT}/bin:${GOPATH}/bin:${PATH}"

RUN GOARCH=${TARGETARCH} CGO_ENABLED=0 go install github.com/go-delve/delve/cmd/dlv@v1.26.3

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

RUN --mount=type=bind,readwrite,source=src/cilium/images/builder,target=/go/src/github.com/cilium/cilium/images/builder \
    ./install-test-formatters.sh

# used to facilitate the verifier tests
COPY --from=cilium-llvm /usr/local/bin/llvm-objcopy /bin/

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
    mv LICENSE.all /tmp/install/linux/amd64/LICENSE.all && \
    mkdir -p /tmp/hubble/linux/amd64 && \
    cd hubble && \
    make GOOS=linux GOARCH=amd64 && \
    mv hubble /tmp/hubble/linux/amd64/hubble

COPY src/cilium/images/cilium/init-container.sh \
     src/cilium/plugins/cilium-cni/install-plugin.sh \
     src/cilium/plugins/cilium-cni/cni-uninstall.sh \
       /tmp/install/linux/amd64

# --------------------------------------------------------------------------------
# https://github.com/cilium/cilium/blob/main/images/cilium/Dockerfile (AS release)
FROM cilium-runtime
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN echo ". /etc/profile.d/bash_completion.sh" >> /etc/bash.bashrc

## When used within the Cilium container, Hubble CLI should target the
## local unix domain socket instead of Hubble Relay.
ENV HUBBLE_SERVER=unix:///var/run/cilium/hubble.sock

COPY --from=cilium-agent-builder /tmp/install/linux/amd64 /
COPY --from=cilium-agent-builder /tmp/hubble/linux/amd64/hubble /usr/bin/hubble
RUN /usr/bin/hubble completion bash > /etc/bash_completion.d/hubble
WORKDIR /home/cilium

ENV INITSYSTEM="SYSTEMD"
ENTRYPOINT ["/usr/bin/cilium-dbg"]
