# ==================== bpftrace-builder image ====================
FROM ghcr.io/cybozu/ubuntu-dev:24.04 AS builder

# https://github.com/bpftrace/bpftrace/blob/master/docker/Dockerfile.ubuntu
RUN apt-get update && apt-get install -y \
    asciidoctor \
    binutils-dev \
    bison \
    build-essential \
    clang \
    cmake \
    flex \
    libbpf-dev \
    libbpfcc-dev \
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
COPY src/bpftrace ./bpftrace
COPY src/libbpf ./libbpf

ENV CC=/usr/bin/clang
ENV CXX=/usr/bin/clang++

# -------------------- Build libbpf --------------------
# https://github.com/libbpf/libbpf?tab=readme-ov-file#building-libbpf
#
# To use the latest libbpf, we need to build it by hand.
WORKDIR /build/libbpf
RUN cd src && \
    mkdir build && \
    BUILD_STATIC_ONLY=y OBJDIR=build make install

# -------------------- Build bpftrace --------------------
# https://github.com/bpftrace/bpftrace/blob/master/INSTALL.md#building-bpftrace
# https://github.com/bpftrace/bpftrace/blob/master/docker/Dockerfile.ubuntu
WORKDIR /build/bpftrace
RUN cmake -B build \
    -DBUILD_TESTING=OFF \
    -DCMAKE_BUILD_TYPE=Release && \
    make -C build -j$(nproc)

# ==================== bpftrace image ====================
FROM ghcr.io/cybozu/ubuntu-debug:24.04
ENV PATH=/opt/bin:"$PATH"

# Minimum requirements found by hand
RUN apt-get update && apt-get install -y --no-install-recommends \
        libbpfcc-dev \
        liblldb-dev \
        libclang-dev \
        libpcap-dev && \
    rm -rf /var/lib/apt/lists/* 

COPY --from=builder /build/bpftrace/build/src/bpftrace /opt/bin/bpftrace
