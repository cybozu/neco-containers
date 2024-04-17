# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG TELEPORT_VERSION=15.1.10
ARG RUST_VERSION=1.71.1
ARG WASM_PACK_VERSION=0.12.1

# Install Nodejs
ARG NODE_VERSION=18.19.1
ENV NODE_URL="https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz"
ENV NODE_PATH="/usr/local/lib/nodejs-linux"
ENV PATH="$PATH:${NODE_PATH}/bin"
RUN mkdir -p ${NODE_PATH} && \
    curl -o /tmp/nodejs.tar.xz -L ${NODE_URL} && \
    tar -xJf /tmp/nodejs.tar.xz -C /usr/local/lib/nodejs-linux --strip-components=1
RUN corepack enable yarn
RUN git clone --depth 1 --branch v${TELEPORT_VERSION} https://github.com/gravitational/teleport && \
    curl https://sh.rustup.rs -sSf | sh -s -- -y && . "$HOME/.cargo/env" && \
    rustup toolchain add ${RUST_VERSION} && rustup default ${RUST_VERSION} && \
    yarn global add wasm-pack@${WASM_PACK_VERSION} && \
    cd teleport && \
    make build/teleport OS=linux

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu-debug:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update && \
    apt-get install -y --no-install-recommends screen tmux && \
    rm -rf /var/lib/apt/lists/* 

COPY --from=build /work/teleport/build/teleport /usr/local/teleport/bin/
COPY --from=build /work/teleport/LICENSE /usr/local/teleport/
COPY update-necocli.sh /etc/profile.d/
COPY loki-env.sh /etc/profile.d/

RUN adduser --disabled-password --gecos "" --uid 10000 cybozu
USER cybozu

ENV PATH=$PATH:/usr/local/teleport/bin

ENTRYPOINT [ "teleport", "start", "-c", "/etc/teleport/teleport.yaml" ]
