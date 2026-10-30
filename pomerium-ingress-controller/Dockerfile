# syntax=docker/dockerfile:1

# Use the same Node.js version as the pomerium container.
ARG NODE_VERSION=24.22.1

FROM scratch AS source
ARG INGRESS_CONTROLLER_VERSION=0.33.4
ARG INGRESS_CONTROLLER_CHECKSUM=0706c45fa28c98ced41aceaa161fac203e1914493899ee90c3a1cc214c991ea3
ADD --checksum=sha256:${INGRESS_CONTROLLER_CHECKSUM} \
    https://github.com/pomerium/ingress-controller/archive/v${INGRESS_CONTROLLER_VERSION}.tar.gz /ingress-controller.tar.gz

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
ARG INGRESS_CONTROLLER_VERSION=0.33.4
ARG INGRESS_CONTROLLER_COMMIT=d53aa35cacac8f193fba174b96753c38a65434b6
ENV GOPROXY=https://golang.flatt.tech

RUN --mount=type=bind,from=node-download,source=/node.tar.gz,target=/tmp/node.tar.gz \
    mkdir -p /opt/node \
 && tar xf /tmp/node.tar.gz -C /opt/node --strip-components=1
ENV PATH="/opt/node/bin:${PATH}"

WORKDIR /work/ingress-controller
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
COPY --from=source /ingress-controller.tar.gz /tmp/ingress-controller.tar.gz
RUN tar zxf /tmp/ingress-controller.tar.gz --strip-components 1 && rm /tmp/ingress-controller.tar.gz

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    go mod download

# The UI of pomerium core is copied from the Go module and embedded into the binary.
# Env var so that `npm ci` skips scripts (esbuild's postinstall is only an optional optimization).
ENV NPM_CONFIG_IGNORE_SCRIPTS=true
RUN make pomerium-ui

# The source tarball has no git metadata, so give the version and the commit explicitly.
RUN make build-go VERSION=v${INGRESS_CONTROLLER_VERSION} GITCOMMIT=${INGRESS_CONTROLLER_COMMIT:0:7}

# ref: https://github.com/pomerium/ingress-controller/blob/v0.33.4/Dockerfile
FROM gcr.io/distroless/base-nossl-debian12:latest@sha256:e35893f8ca1cabfbd60394af94899c20930c4615355adf97c5da660a51b61bca
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/ingress-controller/bin/manager /manager
COPY --from=build /work/ingress-controller/LICENSE /LICENSE
USER 65532:65532

ENTRYPOINT ["/manager"]
