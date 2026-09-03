# syntax=docker/dockerfile:1

# Follow upstream's major/minor version, but refer latest patch version.
# https://github.com/pomerium/pomerium/blob/01af7466fa3b26b5e5369df49f367737d81ad18e/.tool-versions
ARG NODE_VERSION=24.19.0

FROM scratch AS source
ARG POMERIUM_VERSION=0.33.1
ARG POMERIUM_CHECKSUM=fefa6c314d72f8baa17d1736a53840b39fddefe57c3a22ca6bf71a0d3d7091df
ADD --checksum=sha256:${POMERIUM_CHECKSUM} \
    https://github.com/pomerium/pomerium/archive/v${POMERIUM_VERSION}.tar.gz /pomerium.tar.gz

FROM scratch AS node-amd64
ARG NODE_VERSION
ADD --checksum=sha256:f625d97cd707df4ff96254916fbc5ff014f09c09effe5a1e0ca8f6d41a8789d4 \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz /node.tar.gz

FROM scratch AS node-arm64
ARG NODE_VERSION
ADD --checksum=sha256:d28c8a5bf0a808f0ed434a1dce8c54ae98f0371c0bd86ac58abc613f73e6643f \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-arm64.tar.gz /node.tar.gz

# Select the right download stage based on the target architecture
FROM node-${TARGETARCH} AS node-download

FROM ghcr.io/cybozu/golang:1.26.8.2_noble@sha256:e7d0eca1aa1de90597d2c7deeeaa7c93bc7605c51af380ccfc6776c892252a37 AS build

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

# ref: https://github.com/pomerium/pomerium/blob/01af7466fa3b26b5e5369df49f367737d81ad18e/.github/Dockerfile-release
FROM gcr.io/distroless/base-nossl-debian12:latest@sha256:e35893f8ca1cabfbd60394af94899c20930c4615355adf97c5da660a51b61bca
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

WORKDIR /pomerium
COPY --from=build /work/pomerium/bin/* /bin/
COPY --from=build /config.yaml ./config.yaml
COPY --from=build /work/pomerium/LICENSE ./LICENSE
ENTRYPOINT [ "/bin/pomerium" ]
CMD ["-config","/pomerium/config.yaml"]
