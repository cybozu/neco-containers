# cke container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.10-bionic AS build

ARG CKE_TOOLS_VERSION=0.1

RUN mkdir -p /go/src/github.com/cybozu-go/cke
WORKDIR /go/src/github.com/cybozu-go/cke

RUN curl -fsSL -o cke.tar.gz https://github.com/cybozu-go/cke/archive/v${CKE_TOOLS_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f cke.tar.gz \
    && rm -f cke.tar.gz \
    && go get -d -t ./... \
    && go install ./...

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin /usr/local/cke/bin
COPY --from=build /go/src/github.com/cybozu-go/cke/LICENSE /usr/local/cke/LICENSE
COPY install-tools /usr/local/cke/install-tools

ENV PATH=/usr/local/cke/bin:"$PATH"

ENTRYPOINT ["usr/local/cke/bin/cke"]
