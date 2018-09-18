# cke-tools container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG CKE_TOOLS_VERSION=1.0.0
ARG CNI_PLUGIN_VERSION=0.7.1
ARG CNIDIR=/usr/local/cni

WORKDIR /work

RUN curl -fsSL -o cke-tools.tar.gz https://github.com/cybozu-go/cke-tools/archive/v${CKE_TOOLS_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f cke-tools.tar.gz \
    && rm -f cke-tools.tar.gz \
    && go get -d -t ./... \
    && go install ./...

RUN curl -fsSL https://github.com/containernetworking/plugins/archive/v${CNI_PLUGIN_VERSION}.tar.gz | tar zxf - \
    && cd plugins-$CNI_PLUGIN_VERSION \
    && ./build.sh \
    && mkdir -p ${CNIDIR}/bin \
    && mv ./bin/* $CNIDIR/bin

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

ARG CNIDIR=/usr/local/cni

RUN mkdir -p $CNIDIR/bin $CNIDIR/etc/cni/net.d

COPY --from=build /go/bin /usr/local/cke-tools/bin
COPY --from=build /work/LICENSE /usr/local/cke-tools/LICENSE
COPY --from=build /work/scripts/write_file /usr/local/cke-tools/bin
COPY --from=build /work/scripts/make_directories /usr/local/cke-tools/bin
COPY --from=build /work/scripts/install-cni /usr/local/cke-tools/bin
COPY --from=build $CNIDIR/bin/* $CNIDIR/bin/* 
COPY ./99-loopback.conf $CNIDIR/etc/cni/net.d/99-loopback.conf

ENV PATH=/usr/local/cke-tools/bin:"$PATH"
