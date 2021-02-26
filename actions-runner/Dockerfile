FROM quay.io/cybozu/ubuntu:20.04 AS base

ARG ACTIONS_RUNNER_CONTROLLER_VERSION=0.17.0

WORKDIR /work

RUN apt-get update -y && apt-get install -y --no-install-recommends git
RUN git clone -b v${ACTIONS_RUNNER_CONTROLLER_VERSION}  https://github.com/summerwind/actions-runner-controller.git

FROM quay.io/cybozu/ubuntu:20.04

ARG KUBERNETES_VERSION=1.19.7
ARG RUNNER_VERSION=2.274.2
ARG DOCKER_VERSION=19.03.12
ARG DUMB_INIT_VERSION=1.2.2

COPY --from=base /work/actions-runner-controller/runner/entrypoint.sh /
COPY --from=base --chown=runner:docker /work/actions-runner-controller/runner/patched $RUNNER_ASSETS_DIR/patched

RUN apt-get update -y && \
  apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    ca-certificates \
    dnsutils \
    ftp \
    git \
    iproute2 \
    iputils-ping \
    jq \
    libunwind8 \
    locales \
    netcat \
    openssh-client \
    parallel \
    rsync \
    shellcheck \
    sudo \
    telnet \
    time \
    tzdata \
    unzip \
    upx \
    wget \
    zip \
    zstd \
    # this is required by https://github.com/actions/runner
    libyaml-dev \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

RUN curl -L -o /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v${DUMB_INIT_VERSION}/dumb-init_${DUMB_INIT_VERSION}_amd64 \
  && chmod +x /usr/local/bin/dumb-init

RUN curl -L -o docker.tgz https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz \
  && tar zxvf docker.tgz \
  && install -o root -g root -m 755 docker/docker /usr/local/bin/docker \
  && rm -rf docker docker.tgz \
  && adduser --disabled-password --gecos "" --uid 1000 runner \
  && groupadd docker \
  && usermod -aG sudo runner \
  && usermod -aG docker runner \
  && echo "%sudo   ALL=(ALL:ALL) NOPASSWD:ALL" > /etc/sudoers

RUN echo AGENT_TOOLSDIRECTORY=/opt/hostedtoolcache > .env \
  && mkdir /opt/hostedtoolcache \
  && chgrp docker /opt/hostedtoolcache \
  && chmod g+rwx /opt/hostedtoolcache

ARG RUNNER_ASSETS_DIR=/runnertmp
RUN mkdir -p "${RUNNER_ASSETS_DIR}" \
  && cd "${RUNNER_ASSETS_DIR}" \
  && curl -L -o runner.tar.gz https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz \
  && tar xzf ./runner.tar.gz \
  && rm runner.tar.gz \
  && ./bin/installdependencies.sh \
  && mv ./externals ./externalstmp

RUN curl -L https://dl.k8s.io/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl && \
    chmod a+x /usr/local/bin/kubectl

COPY entrypoint_wrapper.sh .

USER 10000:10000
ENTRYPOINT ["/usr/local/bin/dumb-init", "--"]
CMD ["/entrypoint_wrapper.sh"]
