# syntax=docker/dockerfile:1

FROM scratch AS source
ARG POMERIUM_VERSION=0.32.4
ARG POMERIUM_CHECKSUM=bd529ef1ce0793331d74dca674fffe7492f2baf3c09ed5952ca96ffd5bc31b3c
ADD --checksum=sha256:${POMERIUM_CHECKSUM} \
    https://github.com/pomerium/pomerium/archive/v${POMERIUM_VERSION}.tar.gz /pomerium.tar.gz

FROM ghcr.io/cybozu/golang:1.25-noble@sha256:fde615f5be28b80b6a90b3b3bc8276093e7644b1832cbf7fb4ce03b0a703de33 AS build

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

FROM ghcr.io/cybozu/ubuntu:24.04@sha256:ab2735f6893fc167776587097de968c7d66c1cb052326d8cfeb3c4f8cd8bac00
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

WORKDIR /pomerium
COPY --from=build /work/pomerium/bin/* /bin/
COPY --from=build /config.yaml ./config.yaml
COPY --from=build /work/pomerium/LICENSE ./LICENSE
ENTRYPOINT [ "/bin/pomerium" ]
CMD ["-config","/pomerium/config.yaml"]
