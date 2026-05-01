# Stage1: download teleport source code
FROM scratch AS teleport-src
ARG TELEPORT_VERSION=18.7.2
ADD --checksum=sha256:26822b4dbfba8daa672686c235cdff6714c75c9598fdedc8e26ebd20de1aa2ad \
    https://github.com/gravitational/teleport/archive/refs/tags/v${TELEPORT_VERSION}.tar.gz /src.tar.gz

# Stage2: build from source
FROM ghcr.io/cybozu/golang:1.25.9.2_noble@sha256:0aca7f1b3326079d444b33478612421cd6b7b087f50e0a2bd0335173095a1b22 AS build
WORKDIR /work/teleport
RUN --mount=type=bind,from=teleport-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .
RUN make build/teleport OS=linux WEBASSETS_SKIP_BUILD=1

# Stage3: setup runtime container
FROM ghcr.io/cybozu/ubuntu-debug:24.04.20260409@sha256:b457d25587fb885c3d8ef537e573fa19acd253194bce89d3876b45b9ea35ce09
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
