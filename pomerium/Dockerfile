# syntax=docker/dockerfile:1

# Follow upstream's major/minor version, but refer latest patch version.
# https://github.com/pomerium/pomerium/blob/5cd94927fee53f35a2bbd77797cc9524094bff78/.tool-versions
# Exception: Use 24.22.1 for the root certificate store update (eMudhra), which is not available in 24.21.x.
ARG NODE_VERSION=24.22.1

FROM scratch AS source
ARG POMERIUM_VERSION=0.33.4
ARG POMERIUM_CHECKSUM=477bb4909f44b32c73d30655ef21478bd5b403922cf2bbc1549c97286ff2658f
ADD --checksum=sha256:${POMERIUM_CHECKSUM} \
    https://github.com/pomerium/pomerium/archive/v${POMERIUM_VERSION}.tar.gz /pomerium.tar.gz

FROM scratch AS node-amd64
ARG NODE_VERSION
ADD --checksum=sha256:eb6728e77c576a35162c122c45ccadf224d89ff00abc8dae46cc2897c6002d7c \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz /node.tar.gz

FROM scratch AS node-arm64
ARG NODE_VERSION
ADD --checksum=sha256:33c37be234ba8bee041961047b19160f4d7d0f164e8e246f78bdd10feff148b3 \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-arm64.tar.gz /node.tar.gz

# Select the right download stage based on the target architecture
FROM node-${TARGETARCH} AS node-download

FROM ghcr.io/cybozu/golang:1.26.10.2_noble@sha256:8d900a9ee78c02dbe4e2d2e23800c4d21714dfaf8c679935afa668ca0993a0e4 AS build
ENV GOPROXY=https://golang.flatt.tech

RUN --mount=type=bind,from=node-download,source=/node.tar.gz,target=/tmp/node.tar.gz \
    mkdir -p /opt/node \
 && tar xf /tmp/node.tar.gz -C /opt/node --strip-components=1
ENV PATH="/opt/node/bin:${PATH}"

WORKDIR /work/pomerium
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
COPY --from=source /pomerium.tar.gz /tmp/pomerium.tar.gz
RUN tar zxf /tmp/pomerium.tar.gz --strip-components 1 && rm /tmp/pomerium.tar.gz

# Env var so that the npm ci inside `make build-ui` skips scripts too (esbuild's postinstall is only an optional optimization).
ENV NPM_CONFIG_IGNORE_SCRIPTS=true
RUN cd ui; npm ci
RUN make build-ui

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    go mod download
RUN make build-go NAME=pomerium
RUN touch /config.yaml

# ref: https://github.com/pomerium/pomerium/blob/5cd94927fee53f35a2bbd77797cc9524094bff78/.github/Dockerfile-release
FROM gcr.io/distroless/base-nossl-debian12:latest@sha256:e35893f8ca1cabfbd60394af94899c20930c4615355adf97c5da660a51b61bca
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

WORKDIR /pomerium
COPY --from=build /work/pomerium/bin/* /bin/
COPY --from=build /config.yaml ./config.yaml
COPY --from=build /work/pomerium/LICENSE ./LICENSE
ENTRYPOINT [ "/bin/pomerium" ]
CMD ["-config","/pomerium/config.yaml"]
