# hubble container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

COPY TAG /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/cilium/hubble
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL -o hubble.tar.gz "https://github.com/cilium/hubble/archive/v${VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f hubble.tar.gz \
    && rm -f hubble.tar.gz \
    && make clean && make hubble

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/src/github.com/cilium/hubble/hubble /hubble
COPY --from=build /go/src/github.com/cilium/hubble/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/hubble"]
