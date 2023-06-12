# tempo container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.20-jammy AS build

ARG TEMPO_SRCREPO=grafana/tempo
ARG TEMPO_VERSION=2.1.1

WORKDIR /work
RUN git clone --depth=1 -b v${TEMPO_VERSION} https://github.com/${TEMPO_SRCREPO}.git tempo && \
    cd tempo && \
    make tempo

# Stage2: setup runtime containers
FROM quay.io/cybozu/ubuntu:22.04

COPY --from=build /work/tempo/bin/linux/tempo-amd64 /tempo

USER 10000:10000

ENTRYPOINT ["/tempo"]
