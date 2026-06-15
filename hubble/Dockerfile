# hubble container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG HUBBLE_SHA=39037bdd6cb6acb02a241dde33b5ffcf0e8471e2

COPY TAG /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/cilium/hubble
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && git init \
    && git remote add origin https://github.com/cilium/hubble \
    && git fetch --depth 1 origin "refs/tags/v${VERSION}" \
    && test "$(git rev-parse 'FETCH_HEAD^{commit}')" = "${HUBBLE_SHA}" \
    && git checkout --detach "${HUBBLE_SHA}" \
    && make clean && make hubble

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/src/github.com/cilium/hubble/hubble /hubble
COPY --from=build /go/src/github.com/cilium/hubble/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/hubble"]
