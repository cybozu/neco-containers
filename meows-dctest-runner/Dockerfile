ARG MEOWS_VERSION=0.18.1

FROM ghcr.io/cybozu-go/meows-runner:${MEOWS_VERSION}
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

USER root

ENV GO_VERSION=1.23.5
ENV PLACEMAT_VERSION=2.4.7

ENV HOME=/home/actions
ENV GOPATH=${HOME}/go
ENV GOBIN=${GOPATH}/bin
ENV PATH=${GOBIN}:/usr/local/go/bin:${PATH}
ENV NECO_DIR=${GOPATH}/src/github.com/cybozu-go/neco
ENV NECO_APPS_DIR=${GOPATH}/src/github.com/cybozu-private/neco-apps

# Avoid bird post-installation script error
# See https://bird.network.cz/pipermail/bird-users/2019-December/014075.html
COPY include-bird /etc/dpkg/dpkg.cfg.d/

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
            bash-completion \
            bird2 \
            build-essential \
            cloud-utils \
            dbus \
            dnsmasq \
            fakeroot \
            freeipmi-tools \
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
    && rm -rf /var/lib/apt/lists/* \
    && curl -sSLf https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf - \
    && curl -sSfL https://github.com/cybozu-go/placemat/releases/download/v${PLACEMAT_VERSION}/placemat2_${PLACEMAT_VERSION}_amd64.deb -o placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && dpkg -i ./placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && rm ./placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && echo "actions ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers \
    && adduser --disabled-password --gecos "" --uid 10000 actions \
    && mkdir -p /opt/hostedtoolcache \
    && chown -R actions:actions /opt/hostedtoolcache \
    && chown -R actions:actions ${HOME}

USER actions
