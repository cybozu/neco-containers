# cke-tools container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG CKE_TOOLS_VERSION=1.0.1

WORKDIR /work

RUN curl -fsSL -o cke-tools.tar.gz https://github.com/cybozu-go/cke-tools/archive/v${CKE_TOOLS_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f cke-tools.tar.gz \
    && rm -f cke-tools.tar.gz \
    && go get -d -t ./... \
    && go install ./...

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin /usr/local/cke-tools/bin
COPY --from=build /work/LICENSE /usr/local/cke-tools/LICENSE
COPY --from=build /work/scripts/write_file /usr/local/cke-tools/bin
COPY --from=build /work/scripts/make_directories /usr/local/cke-tools/bin

ENV PATH=/usr/local/cke-tools/bin:"$PATH"
