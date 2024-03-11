# tempo container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG TEMPO_SRCREPO=grafana/tempo
ARG TEMPO_VERSION=2.4.0

WORKDIR /work
RUN git clone --depth=1 -b v${TEMPO_VERSION} https://github.com/${TEMPO_SRCREPO}.git tempo && \
    cd tempo && \
    make tempo

# Stage2: setup runtime containers
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/tempo/bin/linux/tempo-amd64 /tempo

USER 10000:10000

ENTRYPOINT ["/tempo"]
