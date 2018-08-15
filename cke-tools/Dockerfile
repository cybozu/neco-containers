# cke-tools container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.10-bionic AS build

ARG CKE_TOOLS_VERSION=0.1

RUN mkdir -p /go/src/github.com/cybozu-go/cke-tools
WORKDIR /go/src/github.com/cybozu-go/cke-tools

RUN curl -fsSL -o cke-tools.tar.gz https://github.com/cybozu-go/cke-tools/archive/v${CKE_TOOLS_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f cke-tools.tar.gz \
    && rm -f cke-tools.tar.gz \
    && go get -d -t ./... \
    && go install ./...

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin /usr/local/cke-tools/bin
COPY --from=build /go/src/github.com/cybozu-go/cke-tools/LICENSE /usr/local/cke-tools/LICENSE

ENV PATH=/usr/local/cke-tools/bin:"$PATH"
