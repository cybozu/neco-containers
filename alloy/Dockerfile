FROM ghcr.io/cybozu/golang:1.24-noble AS build
ARG ALLOY_VERSION=1.8.3
ARG SYSTEMD_VERSION=255

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Dependencies for systemd copied from Build-Depends section written in the control file of groovy systemd package.
# Archived in https://launchpad.net/ubuntu/+archive/primary/+sourcefiles/systemd/246.6-1ubuntu1/systemd_246.6-1ubuntu1.debian.tar.xz
RUN apt-get update && apt-get -y install --no-install-recommends \
    pkg-config \
    xsltproc \
    docbook-xsl \
    docbook-xml \
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
    python3-jinja2 \
    python3-lxml:native \
    python3-pyparsing \
    python3-evdev \
    tzdata \
    libcap2-bin \
    iproute2 \
    zstd

RUN git clone --depth=1 -b v${SYSTEMD_VERSION} https://github.com/systemd/systemd /work/systemd
RUN meson setup /work/systemd/build/ /work/systemd/ && \
    meson install -C /work/systemd/build/

RUN git clone --depth=1 -b v${ALLOY_VERSION} https://github.com/grafana/alloy /work/alloy
COPY force-flush-partial-line.patch /work/
RUN cd alloy && patch -p1 < ../force-flush-partial-line.patch

# we do not add `builtinassets` tag because we do not require alloy ui.
ENV GO_TAGS="netgo promtail_journal_enabled"
RUN make -C /work/alloy USE_CONTAINER=0 alloy

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update && \
    apt-get install --no-install-recommends -y tzdata ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/lib/x86_64-linux-gnu/libsystemd.so.0 /usr/local/systemd/lib/libsystemd.so.0
COPY --from=build /work/systemd/LICENSE.GPL2 /usr/local/systemd/LICENSE.GPL2
COPY --from=build /work/systemd/LICENSE.LGPL2.1 /usr/local/systemd/LICENSE.LGPL2.1
COPY --from=build /work/alloy/build/alloy /bin/alloy
COPY --from=build /work/alloy/LICENSE /usr/local/alloy/LICENSE

ENV LD_LIBRARY_PATH=/usr/local/systemd/lib

RUN mkdir -p /var/lib/alloy/data
RUN chown -R 10000:10000 /var/lib/alloy
RUN chmod -R 770 /var/lib/alloy

USER 10000:10000
ENTRYPOINT ["/bin/alloy"]
ENV ALLOY_DEPLOY_MODE=docker
CMD ["run", "/etc/alloy/config.alloy", "--storage.path=/var/lib/alloy/data"]
