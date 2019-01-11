# coil container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG COIL_VERSION=0.4

WORKDIR /work

RUN curl -fsSL -o coil.tar.gz https://github.com/cybozu-go/coil/archive/v${COIL_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f coil.tar.gz \
    && rm -f coil.tar.gz \
    && CGO_ENABLED=0 GOLDFLAGS="-w -s" go install -mod=vendor ./cmd/hypercoil

WORKDIR /coil

RUN cp /go/bin/hypercoil . \
    && ln -s hypercoil coil \
    && ln -s hypercoil coild \
    && ln -s hypercoil coil-controller \
    && ln -s hypercoil coilctl \
    && ln -s hypercoil coil-installer \
    && cp /work/LICENSE .

# Stage2: setup runtime container
FROM scratch

COPY --from=build /coil /

ENTRYPOINT ["/hypercoil"]
