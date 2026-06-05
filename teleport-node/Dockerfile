# Stage1: download teleport source code
FROM scratch AS teleport-src
ARG TELEPORT_VERSION=18.8.2
ADD --checksum=sha256:4644696c879aa1c48aa2dc7153efd1510b0605de28b72d0a261b950b94e80740 \
    https://github.com/gravitational/teleport/archive/refs/tags/v${TELEPORT_VERSION}.tar.gz /src.tar.gz

# Stage2: build from source
FROM ghcr.io/cybozu/golang:1.25.10.1_noble@sha256:fde615f5be28b80b6a90b3b3bc8276093e7644b1832cbf7fb4ce03b0a703de33 AS build
WORKDIR /work/teleport
RUN --mount=type=bind,from=teleport-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .
RUN make build/teleport OS=linux WEBASSETS_SKIP_BUILD=1

# Stage3: setup runtime container
FROM ghcr.io/cybozu/ubuntu-debug:24.04.20260605@sha256:31a15c78ab811697dce30401a4ce9209f874a46a53bca3664eff583113fd660d
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
