# coil container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG COIL_VERSION=0.2

WORKDIR /work

RUN curl -fsSL -o coil.tar.gz https://github.com/cybozu-go/coil/archive/v${COIL_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f coil.tar.gz \
    && rm -f coil.tar.gz \
    && CGO_ENABLED=0 go install -mod=vendor ./...

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/bin/* /work/LICENSE /
