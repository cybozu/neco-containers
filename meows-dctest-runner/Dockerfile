# syntax=docker/dockerfile:1

ARG MEOWS_RUNNER_TAG=2.335.1-ubuntu22.04-meows0.25.0
ARG MEOWS_RUNNER_CHECKSUM=sha256:a79c4c7ae033e964af29abe5a79014c1d6b7d4d98928b84fb97e43ffe337eb27

FROM scratch AS go-download
ADD --checksum=sha256:1153d3d50e0ac764b447adfe05c2bcf08e889d42a02e0fe0259bd47f6733ad7f \
    https://dl.google.com/go/go1.26.4.linux-amd64.tar.gz /go.tar.gz

FROM scratch AS placemat-download
ADD --checksum=sha256:9f5aa812d00d4fa47dad391e99a85793d6e3b7f7b7b375743d0a56cb8e275186 \
    https://github.com/cybozu-go/placemat/releases/download/v2.5.1/placemat2_2.5.1_amd64.deb /placemat2.deb

FROM scratch AS aqua-installer-download
ADD --checksum=sha256:451028d56959cc738564885b1dbebc2691ea038ffde04e2472e4d486a3591146 \
    https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.5/aqua-installer /aqua-installer

FROM ghcr.io/cybozu-go/meows-runner:${MEOWS_RUNNER_TAG}@${MEOWS_RUNNER_CHECKSUM}
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

USER root

ARG AQUA_VERSION=v2.60.1

ENV HOME=/home/actions
ENV GOPATH=${HOME}/go
ENV GOBIN=${GOPATH}/bin
ENV PATH=/usr/local/aquaproj-aqua/bin:${HOME}/.local/share/aquaproj-aqua/bin:${GOBIN}:/usr/local/go/bin:${PATH}
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

# /home/actions is where a separate volume will be mounted, 
# so install aqua itself under /usr/local/aquaproj-aqua instead.
RUN --mount=type=bind,from=aqua-installer-download,source=/aqua-installer,target=/tmp/aqua-installer,rw \
    chmod +x /tmp/aqua-installer \
    && AQUA_ROOT_DIR=/usr/local/aquaproj-aqua /tmp/aqua-installer -v ${AQUA_VERSION}

USER actions
