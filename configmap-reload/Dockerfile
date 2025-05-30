# configmap-reload container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG CONFIGMAP_RELOAD_VERSION=0.14.0

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/jimmidyson/configmap-reload
RUN curl -fsSL -o configmap-reload.tar.gz "https://github.com/jimmidyson/configmap-reload/archive/v${CONFIGMAP_RELOAD_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f configmap-reload.tar.gz \
    && rm -f configmap-reload.tar.gz \
    && CGO_ENABLED=0 go install --installsuffix cgo -ldflags="-w -s -extldflags '-static'" ./

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/configmap-reload /configmap-reload
COPY --from=build /go/src/github.com/jimmidyson/configmap-reload/LICENSE.txt /LICENSE.txt

USER 10000:10000

ENTRYPOINT ["/configmap-reload"]
