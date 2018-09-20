# cke-tools container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG CKE_TOOLS_VERSION=1.1.0
ARG CNI_PLUGIN_VERSION=0.7.1

WORKDIR /work

RUN curl -fsSL -o cke-tools.tar.gz https://github.com/cybozu-go/cke-tools/archive/v${CKE_TOOLS_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f cke-tools.tar.gz \
    && rm -f cke-tools.tar.gz \
    && go get -d -t ./... \
    && go install ./...

WORKDIR /go/src/github.com/containernetworking/plugins

RUN curl -fsSL https://github.com/containernetworking/plugins/archive/v${CNI_PLUGIN_VERSION}.tar.gz | tar zxf - --strip-components 1 \
    && ./build.sh

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

ARG CNIDIR=/usr/local/cni
ARG PLUGINDIR=/go/src/github.com/containernetworking/plugins

COPY --from=build /go/bin /usr/local/cke-tools/bin
COPY --from=build /work/LICENSE /usr/local/cke-tools/LICENSE
COPY --from=build /work/scripts/write_file /usr/local/cke-tools/bin
COPY --from=build /work/scripts/make_directories /usr/local/cke-tools/bin
COPY --from=build /work/scripts/install-cni /usr/local/cke-tools/bin
COPY --from=build $PLUGINDIR/bin $CNIDIR/bin
COPY --from=build $PLUGINDIR/LICENSE $CNIDIR/LICENSE
COPY ./99-loopback.conf $CNIDIR/etc/cni/net.d/99-loopback.conf

RUN apt-get update \
    && apt-get -y install --no-install-recommends python3 \
    && rm -rf /var/lib/apt/lists/*

ENV PATH=/usr/local/cke-tools/bin:"$PATH"
