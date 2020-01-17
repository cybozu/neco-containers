# cke-tools container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build
ARG CNI_PLUGIN_VERSION=0.8.4

COPY src /work/src
WORKDIR /work/src
RUN go install -mod=vendor ./...

WORKDIR /go/src/github.com/containernetworking/plugins
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/containernetworking/plugins/archive/v${CNI_PLUGIN_VERSION}.tar.gz | tar zxf - --strip-components 1 \
    && ./build_linux.sh

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

ARG CNIDIR=/usr/local/cni
ARG PLUGINDIR=/go/src/github.com/containernetworking/plugins

COPY --from=build /go/bin /usr/local/cke-tools/bin
COPY --from=build /work/src/LICENSE /usr/local/cke-tools/LICENSE
COPY --from=build /work/src/scripts/write_files /usr/local/cke-tools/bin
COPY --from=build /work/src/scripts/make_directories /usr/local/cke-tools/bin
COPY --from=build /work/src/scripts/install-cni /usr/local/cke-tools/bin
COPY --from=build /work/src/scripts/empty-dir /usr/local/cke-tools/bin
COPY --from=build $PLUGINDIR/bin $CNIDIR/bin
COPY --from=build $PLUGINDIR/LICENSE $CNIDIR/LICENSE

RUN apt-get update \
    && apt-get -y install --no-install-recommends python3 \
    && rm -rf /var/lib/apt/lists/*

ENV PATH=/usr/local/cke-tools/bin:"$PATH"
