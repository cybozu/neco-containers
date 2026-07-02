# cadvisor container

# Stage0: download source archive
FROM scratch AS source

ARG CADVISOR_VERSION=v0.57.0
ARG CADVISOR_TARBALL_SHA256=2f700ada18220d6da193e6ecf2d35a52b84e4c38bc20aa08b2e5e3a17f68d1f5

ADD --checksum=sha256:${CADVISOR_TARBALL_SHA256} \
    https://github.com/google/cadvisor/archive/${CADVISOR_VERSION}.tar.gz \
    /cadvisor.tar.gz

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/google/cadvisor
COPY --from=source /cadvisor.tar.gz cadvisor.tar.gz
RUN tar -x -z --strip-components 1 -f cadvisor.tar.gz \
    && cd cmd \
    && CGO_ENABLED=0 go build -tags netgo -ldflags="-w -s" -o cadvisor .

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/src/github.com/google/cadvisor/cmd/cadvisor /cadvisor
COPY --from=build /go/src/github.com/google/cadvisor/LICENSE /LICENSE

EXPOSE 8080

ENTRYPOINT ["/cadvisor", "-logtostderr"]
