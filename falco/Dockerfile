FROM ghcr.io/cybozu/ubuntu-dev:24.04.20261104@sha256:f7e031a6c04507d34e085b97dd7b1dc485996a5e53348d1f5dc73ce6bc159817 AS builder

ARG FALCO_VERSION=0.45.1
ARG FALCO_COMMIT=2869c4b133333006059fae181ecfee7772ab0463 # 0.45.1
ARG BPFTOOL_COMMIT=134c34d987b1a301ffbbc22a51b31502befbeafe # v7.3.0

# reference: https://github.com/falcosecurity/falco/blob/master/.github/workflows/reusable_build_packages.yaml
RUN apt-get update && apt-get install -y --no-install-recommends \
    autoconf \
    automake \
    clang \
    cmake \
    libelf-dev \
    libtool \
    linux-headers-generic \
    llvm \
    m4 \
    pkg-config \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/libbpf/bpftool.git /tmp/bpftool && \
    git -C /tmp/bpftool checkout -q ${BPFTOOL_COMMIT}

WORKDIR /tmp/bpftool
RUN git submodule update --init && \
    make -C src install && \
    rm -rf /tmp/bpftool

WORKDIR /work
RUN curl -fsSL -o /tmp/falco-${FALCO_COMMIT}.tar.gz https://github.com/falcosecurity/falco/archive/${FALCO_COMMIT}.tar.gz && \
    tar -xzf /tmp/falco-${FALCO_COMMIT}.tar.gz --strip-components=1 && \
    rm /tmp/falco-${FALCO_COMMIT}.tar.gz

RUN cmake -B skeleton-build -S . \
    -DUSE_BUNDLED_DEPS=ON \
    -DCREATE_TEST_TARGETS=Off \
    -DFALCO_VERSION=${FALCO_VERSION} \
    && cmake --build skeleton-build --target ProbeSkeleton -j"$(nproc)"

RUN cmake -B build -S . \
    -DCMAKE_BUILD_TYPE=Release \
    -DCPACK_GENERATOR=TGZ \
    -DUSE_BUNDLED_DEPS=On \
    -DFALCO_ETC_DIR=/etc/falco \
    -DMODERN_BPF_SKEL_DIR=/work/skeleton-build/skel_dir \
    -DBUILD_DRIVER=Off \
    -DUSE_JEMALLOC=Off \
    -DUSE_MIMALLOC=Off \
    -DFALCO_VERSION=${FALCO_VERSION} \
    && cmake --build build --target falco -j"$(nproc)"

RUN ARCH=$(uname -m) && \
    cmake --build build --target package && \
    tar -xf "build/falco-${FALCO_VERSION}-${ARCH}.tar.gz" -C /tmp && \
    mv "/tmp/falco-${FALCO_VERSION}-${ARCH}" /install && \
    rm -rf /install/usr/src/falco-* \
           /install/usr/bin/falcoctl \
           /install/etc/falcoctl

FROM ghcr.io/cybozu/ubuntu:24.04.20261104@sha256:d352ea0ca39845d0c724bcb974826bb530ea3879e7d1637e8b3d27797adcfa0d
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /install/ /
# Change the falco config within the container to enable ISO 8601 output.
COPY falco.iso8601_timeformat.yaml /etc/falco/config.d/

CMD ["/usr/bin/falco"]
