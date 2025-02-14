# ==================== bpftrace-builder image ====================
FROM ghcr.io/cybozu/ubuntu-dev:24.04 AS builder

# https://github.com/bpftrace/bpftrace/blob/master/docker/Dockerfile.ubuntu
# Added: libcurl4-openssl-dev and libedit-dev
#   - possibly because of the difference in base image
# Removed: libbpf-dev libbpfcc-dev
#   - to use locally-built bcc and libbpf
RUN apt-get update && apt-get install -y \
    asciidoctor \
    binutils-dev \
    bison \
    build-essential \
    clang \
    cmake \
    flex \
    libcereal-dev \
    libelf-dev \
    libiberty-dev \
    libpcap-dev \
    llvm-dev \
    liblldb-dev \
    libclang-dev \
    systemtap-sdt-dev \
    zlib1g-dev \
    libcurl4-openssl-dev \
    libedit-dev

WORKDIR /build
COPY Makefile .
COPY src/bcc ./bcc
COPY src/bpftrace ./bpftrace
COPY src/libbpf ./libbpf

ENV CC=/usr/bin/clang
ENV CXX=/usr/bin/clang++

# -------------------- Build bcc --------------------
# https://github.com/iovisor/bcc/blob/master/INSTALL.md#ubuntu---source
#
# bcc and bpftrace are closely-related tools and bpftrace depends on bcc.
# For their difference, see:
# - Linux Extended BPF (eBPF) Tracing Tools
#   https://www.brendangregg.com/ebpf.html
WORKDIR /build/bcc

# Removed: arping, netperf, and iperf
#   - they are installed in upstream images to run examples
# Please see the upstream instruction for further info.
RUN apt-get install -y --no-install-recommends \
        zip \
        bison \
        build-essential \
        cmake \
        flex \
        git \
        libedit-dev \
        libllvm18 \
        llvm-18-dev \
        libclang-18-dev \
        python3 \
        zlib1g-dev \
        libelf-dev \
        libfl-dev \
        python3-setuptools \
        liblzma-dev \
        libdebuginfod-dev \
        libpolly-18-dev

# Missing -DREVISION causes an error, although it is not documented
# ref. https://github.com/iovisor/bcc/blob/master/cmake/version.cmake
RUN mkdir build && \
    cmake -B build -DREVISION="$(make --no-print-directory -C '..' bcc-version)" && \
    make -C build -j$(nproc) && \
    make -C build install

# -------------------- Build libbpf --------------------
# https://github.com/libbpf/libbpf?tab=readme-ov-file#building-libbpf
#
# To use the latest libbpf, we need to build it by hand.
WORKDIR /build/libbpf
RUN cd src && \
    mkdir build && \
    OBJDIR=build LIBDIR=/usr/local/lib make install

# -------------------- Build bpftrace --------------------
# https://github.com/bpftrace/bpftrace/blob/master/INSTALL.md#building-bpftrace
# https://github.com/bpftrace/bpftrace/blob/master/docker/Dockerfile.ubuntu
WORKDIR /build/bpftrace
RUN cmake -B build \
    -DBUILD_TESTING=OFF \
    -DCMAKE_BUILD_TYPE=Release && \
    make -C build -j$(nproc)

# ==================== bpftrace image ====================
# We cannot use ghcr.io/cybozu/ubuntu-debug, because it has libbpf1 installed
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ENV PATH=/opt/bin:"$PATH"

RUN apt-get update && apt-get install -y --no-install-recommends \
# Minimum requirements found by hand
        binutils-dev \
        liblldb-dev \
        libclang-dev \
        libdebuginfod-dev \
        libelf-dev \
        libpcap-dev \
# Install tools
        bsdmainutils \
        file \
        jq \
        vim && \
    rm -rf /var/lib/apt/lists/* 

COPY --from=builder /usr/lib/x86_64-linux-gnu/libbcc* /usr/lib/x86_64-linux-gnu/
COPY --from=builder /usr/local/lib/libbpf* /usr/local/lib/
COPY --from=builder /build/bpftrace/build/src/bpftrace /opt/bin/bpftrace
