ARG MEOWS_VERSION=0.9.1

FROM quay.io/cybozu/meows-runner:${MEOWS_VERSION}

USER root

ENV GO_VERSION=1.18.4
ENV PLACEMAT_VERSION=2.3.0

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
    && apt-get install -y software-properties-common \
    && add-apt-repository -y ppa:smoser/swtpm \
    && apt-get update \
    && apt-get -y install --no-install-recommends \
            git \
            build-essential \
            less \
            wget \
            systemd-container \
            lldpd \
            qemu \
            qemu-kvm \
            socat \
            picocom \
            swtpm \
            cloud-utils \
            bird2 \
            squid \
            dnsmasq \
            xauth \
            bash-completion \
            dbus \
            jq \
            libgpgme11 \
            freeipmi-tools \
            unzip \
            fakeroot \
            time \
            kmod \
            iptables \
            iproute2 \
            openssh-client \
            sudo \
    && rm -rf /var/lib/apt/lists/* \
    && curl -sSLf https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf - \
    && curl -sfL https://github.com/cybozu-go/placemat/releases/download/v${PLACEMAT_VERSION}/placemat2_${PLACEMAT_VERSION}_amd64.deb -o placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && dpkg -i ./placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && rm ./placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && echo "actions ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers \
    && adduser --disabled-password --gecos "" --uid 10000 actions \
    && mkdir -p /opt/hostedtoolcache \
    && chown -R actions:actions /opt/hostedtoolcache \
    && chown -R actions:actions ${HOME}

COPY --chown=actions:actions dctest-bootstrap /usr/local/bin
COPY --chown=actions:actions neco-bootstrap /usr/local/bin
COPY --chown=actions:actions neco-apps-bootstrap /usr/local/bin

RUN chmod +x /usr/local/bin/dctest-bootstrap \
    && chmod +x /usr/local/bin/neco-bootstrap \
    && chmod +x /usr/local/bin/neco-apps-bootstrap

USER actions
