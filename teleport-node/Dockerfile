# Stage1: build from source
FROM quay.io/cybozu/golang:1.19-focal AS build

ARG TELEPORT_VERSION=11.3.3
RUN git clone --depth 1 --branch v${TELEPORT_VERSION} https://github.com/gravitational/teleport && \
    cd teleport && \
    make build/teleport OS=linux

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu-debug:20.04

COPY --from=build /work/teleport/build/teleport /usr/local/teleport/bin/
COPY --from=build /work/teleport/LICENSE /usr/local/teleport/
COPY update-necocli.sh /etc/profile.d/
COPY loki-env.sh /etc/profile.d/

RUN adduser --disabled-password --gecos "" --uid 10000 cybozu
USER cybozu

ENV PATH=$PATH:/usr/local/teleport/bin

ENTRYPOINT [ "teleport", "start", "-c", "/etc/teleport/teleport.yaml" ]
