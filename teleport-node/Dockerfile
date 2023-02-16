# Stage1: build from source
FROM quay.io/cybozu/golang:1.19-focal AS build

ARG TELEPORT_VERSION=11.3.3
# Install Nodejs
ARG NODE_VERSION=16.18.1
ENV NODE_URL="https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz"
ENV NODE_PATH="/usr/local/lib/nodejs-linux"
ENV PATH="$PATH:${NODE_PATH}/bin"
RUN mkdir -p ${NODE_PATH} && \
    curl -o /tmp/nodejs.tar.xz -L ${NODE_URL} && \
    tar -xJf /tmp/nodejs.tar.xz -C /usr/local/lib/nodejs-linux --strip-components=1
RUN corepack enable yarn

RUN git clone --depth 1 --branch v${TELEPORT_VERSION} https://github.com/gravitational/teleport && \
    cd teleport && \
    make build/teleport OS=linux

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu-debug:20.04

COPY --from=build /work/teleport/build/teleport /usr/local/teleport/bin/
COPY --from=build /work/teleport/LICENSE /usr/local/teleport/
COPY update-necocli.sh /etc/profile.d/
COPY loki-env.sh /etc/profile.d/

RUN adduser --disabled-password --gecos "" --uid 10000 cybozu
USER cybozu

ENV PATH=$PATH:/usr/local/teleport/bin

ENTRYPOINT [ "teleport", "start", "-c", "/etc/teleport/teleport.yaml" ]
