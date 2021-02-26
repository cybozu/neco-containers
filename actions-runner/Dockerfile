FROM quay.io/cybozu/ubuntu:20.04 AS base

ARG ACTIONS_RUNNER_CONTROLLER_VERSION=0.17.0

WORKDIR /work

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -y \
  && apt-get install -y --no-install-recommends \
  git
RUN git clone -b v${ACTIONS_RUNNER_CONTROLLER_VERSION}  https://github.com/summerwind/actions-runner-controller.git

FROM quay.io/cybozu/ubuntu:20.04

ARG ARCH=amd64
ARG RUNNER_VERSION=2.274.2
ARG DOCKER_VERSION=19.03.12

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -y \
  && apt-get install -y --no-install-recommends \
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
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

RUN curl -L -o /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 \
  && chmod +x /usr/local/bin/dumb-init

# Docker download supports arm64 as aarch64 & amd64 as x86_64
RUN curl -L -o docker.tgz https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz \
  && tar zxvf docker.tgz \
  && install -o root -g root -m 755 docker/docker /usr/local/bin/docker \
  && rm -rf docker docker.tgz \
  && adduser --disabled-password --gecos "" --uid 1000 runner \
  && groupadd docker \
  && usermod -aG sudo runner \
  && usermod -aG docker runner \
  && echo "%sudo   ALL=(ALL:ALL) NOPASSWD:ALL" > /etc/sudoers

ENV RUNNER_ASSETS_DIR=/runnertmp

# Runner download supports amd64 as x64. Externalstmp is needed for making mount points work inside DinD.
#
# libyaml-dev is required for ruby/setup-ruby action.
# It is installed after installdependencies.sh and before removing /var/lib/apt/lists
# to avoid rerunning apt-update on its own.
RUN mkdir -p "$RUNNER_ASSETS_DIR" \
  && cd "$RUNNER_ASSETS_DIR" \
  && curl -L -o runner.tar.gz https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz \
  && tar xzf ./runner.tar.gz \
  && rm runner.tar.gz \
  && ./bin/installdependencies.sh \
  && mv ./externals ./externalstmp \
  && apt-get install -y libyaml-dev \
  && rm -rf /var/lib/apt/lists/*

RUN echo AGENT_TOOLSDIRECTORY=/opt/hostedtoolcache > .env \
  && mkdir /opt/hostedtoolcache \
  && chgrp docker /opt/hostedtoolcache \
  && chmod g+rwx /opt/hostedtoolcache

COPY --from=base /work/actions-runner-controller/runner/entrypoint.sh /
COPY --from=base --chown=runner:docker /work/actions-runner-controller/runner/patched $RUNNER_ASSETS_DIR/patched

ARG KUBERNETES_VERSION

USER root
RUN curl -L https://dl.k8s.io/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl && \
    chmod 755 /usr/local/bin/kubectl

USER runner
COPY entrypoint_wrapper.sh .

ENTRYPOINT ["/usr/local/bin/dumb-init", "--"]
CMD ["/entrypoint_wrapper.sh"]
