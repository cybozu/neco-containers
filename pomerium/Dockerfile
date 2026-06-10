# syntax=docker/dockerfile:1

FROM scratch AS source
ARG POMERIUM_VERSION=0.32.7
ARG POMERIUM_CHECKSUM=7a3fd6d8b3b1ff49645a82da1a83a6713afb1b631cf45bb2dab56eeb0c2dfaeb
ADD --checksum=sha256:${POMERIUM_CHECKSUM} \
    https://github.com/pomerium/pomerium/archive/v${POMERIUM_VERSION}.tar.gz /pomerium.tar.gz

FROM ghcr.io/cybozu/golang:1.25.11.1_noble@sha256:cdfa4c5b7ceb83fe059fb193182d7fc72bf0a3375dd5d9a00ff9d9dc7f5a857d AS build

WORKDIR /work/pomerium
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
COPY --from=source /pomerium.tar.gz /tmp/pomerium.tar.gz
RUN tar zxf /tmp/pomerium.tar.gz --strip-components 1 && rm /tmp/pomerium.tar.gz

ARG NODE_VERSION=22.x

RUN apt-get update \
    && apt-get install -y --no-install-recommends gnupg zip \
    && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /usr/share/keyrings/nodesource.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/nodesource.gpg] https://deb.nodesource.com/node_${NODE_VERSION} nodistro main" \
        | tee /etc/apt/sources.list.d/nodesource.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*
RUN cd ui; npm ci
RUN make build-ui

RUN go mod download
RUN make build-go NAME=pomerium
RUN touch /config.yaml

FROM ghcr.io/cybozu/ubuntu:24.04.20260604@sha256:c7c519ae7504df61e01495635ebda9c8698f2d221cd3fcf6e9500ad97b9d2421
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

WORKDIR /pomerium
COPY --from=build /work/pomerium/bin/* /bin/
COPY --from=build /config.yaml ./config.yaml
COPY --from=build /work/pomerium/LICENSE ./LICENSE
ENTRYPOINT [ "/bin/pomerium" ]
CMD ["-config","/pomerium/config.yaml"]
