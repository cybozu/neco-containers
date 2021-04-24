# Fluent Bit container image

# Stage1: build from source
FROM quay.io/cybozu/ubuntu-dev:20.04 AS builder

ARG FLUENT_BIT_VERSION=1.7.4
ARG SYSTEMD_VERSION=247

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN apt-get update && apt-get -y install --no-install-recommends \
    pkg-config \
    xsltproc \
    docbook-xsl \
    docbook-xml \
    m4 \
    meson \
    gettext \
    gperf \
    gnu-efi \
    libcap-dev \
    libpam0g-dev \
    libapparmor-dev \
    libidn2-dev \
    libiptc-dev \
    libaudit-dev \
    libdbus-1-dev \
    libcryptsetup-dev \
    libselinux1-dev \
    libacl1-dev \
    liblzma-dev \
    liblz4-dev \
    liblz4-tool \
    libbz2-dev \
    zlib1g-dev \
    libcurl4-gnutls-dev \
    libmicrohttpd-dev \
    libgnutls28-dev \
    libpcre2-dev \
    libgcrypt20-dev \
    libkmod-dev \
    libblkid-dev \
    libmount-dev \
    libseccomp-dev \
    libdw-dev \
    libpolkit-gobject-1-dev \
    libzstd-dev \
    linux-base \
    acl \
    python3:native \
    python3-lxml:native \
    python3-pyparsing \
    python3-evdev \
    tzdata \
    libcap2-bin \
    iproute2 \
    zstd \
    cmake \
    flex \
    bison \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /work \
    && git clone --depth=1 -b v${SYSTEMD_VERSION} https://github.com/systemd/systemd /work/systemd \
    && meson setup /work/systemd/build/ /work/systemd/ \
    && meson configure -Drootlibdir=/usr/lib/$(dpkg-architecture -q DEB_BUILD_MULTIARCH) /work/systemd/build/ \
    && meson install -C /work/systemd/build/

RUN mkdir -p /work/fluent-bit \
    && curl -sSLf https://github.com/fluent/fluent-bit/archive/v${FLUENT_BIT_VERSION}.tar.gz | \
         tar -x -f - -z --strip-components=1 -C /work/fluent-bit \
    && cd /work/fluent-bit/build \
    && cmake .. \
    && make -j $(nproc) \
    && strip bin/fluent-bit \
    && mkdir -p /fluent-bit/bin \
    && cp bin/fluent-bit /fluent-bit/bin/ \
    && mkdir -p /fluent-bit/etc \
    && cd /work/fluent-bit \
    && cp -a conf/fluent-bit.conf \
        conf/parsers.conf \
        conf/parsers_ambassador.conf \
        conf/parsers_java.conf \
        conf/parsers_extra.conf \
        conf/parsers_openstack.conf \
        conf/parsers_cinder.conf \
        conf/plugins.conf \
        /fluent-bit/etc/ \
    && cp LICENSE /fluent-bit/

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

# libsystemd depends on libcap2.
RUN apt-get update \
    && apt-get -y install --no-install-recommends \
       libcap2 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /fluent-bit /fluent-bit
COPY --from=builder /lib/x86_64-linux-gnu/libsystemd.so.0.30.0 /lib/x86_64-linux-gnu/libsystemd.so.0.30.0

RUN ln -snf /lib/x86_64-linux-gnu/libsystemd.so.0.30.0 /lib/x86_64-linux-gnu/libsystemd.so.0

EXPOSE 2020

ENTRYPOINT ["/fluent-bit/bin/fluent-bit"]
CMD ["-q", "-c", "/fluent-bit/etc/fluent-bit.conf"]
