# Stage1: download teleport source code
FROM scratch AS teleport-src
ARG TELEPORT_VERSION=18.11.1
ADD --checksum=sha256:a95bddd445dc3f93d91493b48e73d39959bbc4d4d00cd93fea34b3d02c3ba1da \
    https://github.com/gravitational/teleport/archive/refs/tags/v${TELEPORT_VERSION}.tar.gz /src.tar.gz

# Stage2: build from source
FROM ghcr.io/cybozu/golang:1.25.14.1_noble@sha256:1b449b9c9d5f41d55fb70370466555a96588c5ec83976b8fef669f2f05a8ed16 AS build
ENV GOPROXY=https://golang.flatt.tech
WORKDIR /work/teleport
RUN --mount=type=bind,from=teleport-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    make build/teleport OS=linux WEBASSETS_SKIP_BUILD=1

# Stage3: setup runtime container
FROM ghcr.io/cybozu/ubuntu-debug:24.04.20261022@sha256:ca40efcb3c9675220306fa72d6c1f6ab00d906e95d696f833fc0e95fd96f68fb
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
