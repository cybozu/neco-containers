# syntax=docker/dockerfile:1

# Follow upstream's major/minor version, but refer latest patch version.
# https://github.com/pomerium/pomerium/blob/593eb81c7e5bdbe6071a30d330f374967869577f/.tool-versions
ARG NODE_VERSION=22.22.3

FROM scratch AS source
ARG POMERIUM_VERSION=0.32.7
ARG POMERIUM_CHECKSUM=7a3fd6d8b3b1ff49645a82da1a83a6713afb1b631cf45bb2dab56eeb0c2dfaeb
ADD --checksum=sha256:${POMERIUM_CHECKSUM} \
    https://github.com/pomerium/pomerium/archive/v${POMERIUM_VERSION}.tar.gz /pomerium.tar.gz

FROM scratch AS node-amd64
ARG NODE_VERSION
ADD --checksum=sha256:c7a10d6816da8eaaa7534dd73c71c6e2b2c391dbbf845e364902d156615dd1b8 \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz /node.tar.gz

FROM scratch AS node-arm64
ARG NODE_VERSION
ADD --checksum=sha256:cc8bc82b2dd0b595c3b95a4c3c9c8c350907cff011afbdee3d1379e812e1e3e3 \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-arm64.tar.gz /node.tar.gz

# Select the right download stage based on the target architecture
FROM node-${TARGETARCH} AS node-download

FROM ghcr.io/cybozu/golang:1.25.11.1_noble@sha256:cdfa4c5b7ceb83fe059fb193182d7fc72bf0a3375dd5d9a00ff9d9dc7f5a857d AS build

RUN --mount=type=bind,from=node-download,source=/node.tar.gz,target=/tmp/node.tar.gz \
    mkdir -p /opt/node \
 && tar xf /tmp/node.tar.gz -C /opt/node --strip-components=1
ENV PATH="/opt/node/bin:${PATH}"

WORKDIR /work/pomerium
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
COPY --from=source /pomerium.tar.gz /tmp/pomerium.tar.gz
RUN tar zxf /tmp/pomerium.tar.gz --strip-components 1 && rm /tmp/pomerium.tar.gz

RUN cd ui; npm ci
RUN make build-ui

RUN go mod download
RUN make build-go NAME=pomerium
RUN touch /config.yaml

FROM ghcr.io/cybozu/ubuntu:24.04.20260608@sha256:2137d223a483a2870dae87054a21314a69c6d8b9583a9a4ab25ea6e87b178b4a
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

WORKDIR /pomerium
COPY --from=build /work/pomerium/bin/* /bin/
COPY --from=build /config.yaml ./config.yaml
COPY --from=build /work/pomerium/LICENSE ./LICENSE
ENTRYPOINT [ "/bin/pomerium" ]
CMD ["-config","/pomerium/config.yaml"]
