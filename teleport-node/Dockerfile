# Stage1: download teleport source code
FROM scratch AS teleport-src
ARG TELEPORT_VERSION=18.7.6
ADD --checksum=sha256:30fb59382ba7e52f799c4ac3f950b3d7a4fbdfba2a826c8789c4ec4ac8e08f8c \
    https://github.com/gravitational/teleport/archive/refs/tags/v${TELEPORT_VERSION}.tar.gz /src.tar.gz

# Stage2: build from source
FROM ghcr.io/cybozu/golang:1.25.9.2_noble@sha256:0aca7f1b3326079d444b33478612421cd6b7b087f50e0a2bd0335173095a1b22 AS build
WORKDIR /work/teleport
RUN --mount=type=bind,from=teleport-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .
RUN make build/teleport OS=linux WEBASSETS_SKIP_BUILD=1

# Stage3: setup runtime container
FROM ghcr.io/cybozu/ubuntu-debug:24.04.20260507@sha256:098abdd87057f7a7bf5d2a1a71d29a1b4c2b49f685ae60efd8fdade72a8bb92f
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
