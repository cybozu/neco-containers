# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.23-jammy AS build

ARG TELEPORT_VERSION=16.5.6

RUN git clone --depth 1 --branch v${TELEPORT_VERSION} https://github.com/gravitational/teleport && \
    cd teleport && \
    make build/teleport OS=linux WEBASSETS_SKIP_BUILD=1

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
COPY hubble-env.sh /etc/profile.d/
COPY bash-env.sh /etc/profile.d/

USER cybozu

ENV PATH=$PATH:/usr/local/teleport/bin

ENTRYPOINT [ "teleport", "start", "-c", "/etc/teleport/teleport.yaml" ]
