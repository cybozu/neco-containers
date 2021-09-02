FROM quay.io/cybozu/meows-runner:0.3.1

USER root

ENV GO_VERSION=1.16.6
ENV PLACEMAT_VERSION=2.0.5

ENV HOME=/home/actions
ENV GOPATH=${HOME}/go
ENV GOBIN=${GOPATH}/bin
ENV PATH=${GOBIN}:/usr/local/go/bin:${PATH}
ENV NECO_DIR=${GOPATH}/src/github.com/cybozu-go/neco
ENV NECO_APPS_DIR=${GOPATH}/src/github.com/cybozu-go/neco-apps

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
            chrony \
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
            openssh-client \
            sudo \
            hub \
    && rm -rf /var/lib/apt/lists/* \
    && curl -sSLf https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf - \
    && curl -sfL https://github.com/cybozu-go/placemat/releases/download/v${PLACEMAT_VERSION}/placemat2_${PLACEMAT_VERSION}_amd64.deb -o placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && dpkg -i ./placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && rm ./placemat2_${PLACEMAT_VERSION}_amd64.deb \
    && git clone -b ${NECO_BRANCH:-release} https://github.com/cybozu-go/neco.git ${NECO_DIR} \
    && make -C ${NECO_DIR}/dctest setup \
    && git clone -b ${NECO_APPS_BRANCH:-release} https://github.com/cybozu-go/neco-apps.git ${NECO_APPS_DIR} \
    && make -C ${NECO_APPS_DIR}/test setup \
    && echo "actions ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers \
    && adduser --disabled-password --gecos "" --uid 10000 actions \
    && mkdir -p /opt/hostedtoolcache \
    && chown -R actions:actions /opt/hostedtoolcache \
    && chown -R actions:actions ${HOME}

COPY bootstrap.sh ${HOME}

USER actions
