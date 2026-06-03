# syntax=docker/dockerfile:1

FROM scratch AS go-download
ADD --checksum=sha256:2b2cfc7148493da5e73981bffbf3353af381d5f93e789c82c79aff64962eb556 \
    https://dl.google.com/go/go1.26.3.linux-amd64.tar.gz /go.tar.gz

FROM scratch AS placemat-download
ADD --checksum=sha256:eb32893ed824d8d009cd7214283c02b9652a460fa3a7184c47e7adbcc0f32cd2 \
    https://github.com/cybozu-go/placemat/releases/download/v2.4.11/placemat2_2.4.11_amd64.deb /placemat2.deb

FROM ghcr.io/cybozu-go/meows-runner:2.334.0-ubuntu22.04-meows0.24.0@sha256:bb11e4c0b433f8650a36071c60ebf833da5667f45f477bba40f18ef78001bae8
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

USER root

ENV HOME=/home/actions
ENV GOPATH=${HOME}/go
ENV GOBIN=${GOPATH}/bin
ENV PATH=${GOBIN}:/usr/local/go/bin:${PATH}
ENV NECO_DIR=${GOPATH}/src/github.com/cybozu-go/neco
ENV NECO_APPS_DIR=${GOPATH}/src/github.com/cybozu-private/neco-apps

# Disable gh telemetry: https://cli.github.com/telemetry
ENV GH_TELEMETRY=false
ENV DO_NOT_TRACK=true

# Avoid bird post-installation script error
# See https://bird.network.cz/pipermail/bird-users/2019-December/014075.html
COPY include-bird /etc/dpkg/dpkg.cfg.d/

RUN curl -sSLf https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get -y install --no-install-recommends \
            bash-completion \
            bird2 \
            build-essential \
            cloud-utils \
            dbus \
            dnsmasq \
            fakeroot \
            freeipmi-tools \
            gh \
            git \
            iproute2 \
            iptables \
            jq \
            kmod \
            less \
            libgpgme11 \
            lldpd \
            lsb-release \
            openssh-client \
            picocom \
            qemu \
            qemu-kvm \
            socat \
            squid \
            sudo \
            swtpm \
            systemd-container \
            time \
            unzip \
            wget \
            xauth \
    && rm -rf /var/lib/apt/lists/*

RUN --mount=type=bind,from=go-download,source=/go.tar.gz,target=/tmp/go.tar.gz,ro \
    tar -C /usr/local -xzf /tmp/go.tar.gz

RUN --mount=type=bind,from=placemat-download,source=/placemat2.deb,target=/tmp/placemat2.deb,ro \
    dpkg -i /tmp/placemat2.deb \
    && echo "actions ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers \
    && adduser --disabled-password --gecos "" --uid 10000 actions \
    && mkdir -p /opt/hostedtoolcache \
    && chown -R actions:actions /opt/hostedtoolcache \
    && chown -R actions:actions ${HOME}

USER actions
