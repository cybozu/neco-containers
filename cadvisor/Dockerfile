# cadvisor container

# Stage0: download source archive
FROM scratch AS source

ARG CADVISOR_VERSION=v0.60.3
ARG CADVISOR_TARBALL_SHA256=4d43907f4f8d87bd4c59f3c829ca2724283f52ec67e671142c6246e66be15393

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
