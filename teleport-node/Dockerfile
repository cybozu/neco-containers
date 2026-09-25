# Stage1: download teleport source code
FROM scratch AS teleport-src
ARG TELEPORT_VERSION=18.10.0
ADD --checksum=sha256:75a6018ebb6e7e5ffc9f74ee6a5352dc1d5a3390f5617945cbef4f1a45f614e5 \
    https://github.com/gravitational/teleport/archive/refs/tags/v${TELEPORT_VERSION}.tar.gz /src.tar.gz

# Stage2: build from source
FROM ghcr.io/cybozu/golang:1.27.1.1_noble@sha256:e38fe3b72f61d034394ee2c2592d41fa753226718bbc111bb6bb9a21601f0859 AS build
WORKDIR /work/teleport
RUN --mount=type=bind,from=teleport-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .
RUN make build/teleport OS=linux WEBASSETS_SKIP_BUILD=1

# Stage3: setup runtime container
FROM ghcr.io/cybozu/ubuntu-debug:24.04.20260925@sha256:d6e4251ccefc7b3cba0af18549d98d4338ec7bfeab284d260e9000ee980f1216
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
