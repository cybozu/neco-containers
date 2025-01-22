FROM ghcr.io/cybozu/golang:1.23-noble AS build
COPY . /work/src
WORKDIR /work/src
RUN make build GO_BUILD_OPT='-ldflags="-w -s"'

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/bin/lbip-manager /
ENTRYPOINT ["/lbip-manager"]
