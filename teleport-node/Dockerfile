# Stage1: build from source
FROM quay.io/cybozu/golang:1.15-focal AS build

ARG TELEPORT_VERSION=6.0.0
RUN git clone --depth 1 --branch v${TELEPORT_VERSION} https://github.com/gravitational/teleport && \
    cd teleport && \
    make build/teleport OS=linux

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu-debug:20.04

COPY --from=build /work/teleport/build/teleport /usr/local/bin/
COPY update-necocli.sh /etc/profile.d/

RUN adduser --disabled-password --gecos "" --uid 10000 cybozu
RUN chown cybozu:cybozu /home/cybozu
USER cybozu

ENTRYPOINT [ "teleport", "start", "-c", "/etc/teleport/teleport.yaml" ]
