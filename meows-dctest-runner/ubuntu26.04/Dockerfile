# syntax=docker/dockerfile:1

FROM ghcr.io/cybozu-go/meows-runner:2.338.2-ubuntu24.04-meows0.30.0@sha256:b79eb0984707d2eabb55a1e7474e94352f7c81f4361ad3cc1eea7b820a3eba36 AS runner

FROM scratch AS go-download
ADD --checksum=sha256:ccfe7ae822d2c2f41f4fe7c956199967a4dac22d0ee0d2b2bfdb440aab3b20c5 \
    https://dl.google.com/go/go1.26.10.linux-amd64.tar.gz /go.tar.gz

FROM scratch AS placemat-download
ADD --checksum=sha256:655b84737ea48a59f348ad5f32c3c2f9c04ecfdf44a6a43da9ad34311c4e3eed \
    https://github.com/cybozu-go/placemat/releases/download/v2.7.2/placemat2_2.7.2_amd64.deb /placemat2.deb

FROM scratch AS aqua-installer-download
ADD --checksum=sha256:c0acde4802a1edd8b8201fd4cc1f0e52700cd4feb875f9c9793381b5b0ca2065 \
    https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.1.1/aqua-installer /aqua-installer

FROM ghcr.io/cybozu/ubuntu:26.04.20261111@sha256:4c899d982515115b4de5821fbd6a2208fec224244b2d6e9bb1afc396d199632f
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

USER root

ARG AQUA_VERSION=v2.65.2

ENV DEBIAN_FRONTEND=noninteractive
ENV HOME=/home/actions
ENV GOPATH=${HOME}/go
ENV GOBIN=${GOPATH}/bin
ENV PATH=/usr/local/aquaproj-aqua/bin:${HOME}/.local/share/aquaproj-aqua/bin:${GOBIN}:/usr/local/go/bin:${PATH}
ENV NECO_DIR=${GOPATH}/src/github.com/cybozu-go/neco
ENV NECO_APPS_DIR=${GOPATH}/src/github.com/cybozu-private/neco-apps
ENV AGENT_TOOLSDIRECTORY=/opt/hostedtoolcache
ENV GH_TELEMETRY=false
ENV DO_NOT_TRACK=true

COPY include-bird /etc/dpkg/dpkg.cfg.d/
COPY --from=runner --chown=10000:10000 /runner /runner
COPY --from=runner /usr/local/bin/meows /usr/local/bin/job-started /usr/local/bin/entrypoint /usr/local/bin/job-cancelled /usr/local/bin/job-failure /usr/local/bin/job-success /usr/local/bin/

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
            libgpgme45 \
            libyaml-dev \
            lldpd \
            lsb-release \
            openssh-client \
            picocom \
            qemu-system-x86 \
            qemu-utils \
            socat \
            squid \
            sudo \
            swtpm \
            systemd-container \
            time \
            unzip \
            wget \
            xauth \
    && /runner/bin/installdependencies.sh \
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

RUN --mount=type=bind,from=aqua-installer-download,source=/aqua-installer,target=/tmp/aqua-installer,rw \
    chmod +x /tmp/aqua-installer \
    && AQUA_ROOT_DIR=/usr/local/aquaproj-aqua /tmp/aqua-installer -v ${AQUA_VERSION}

USER actions
CMD ["/usr/local/bin/entrypoint"]
