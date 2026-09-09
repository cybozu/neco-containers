ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.26.7.1_noble@sha256:678dffc420bde6361513f3c63a5c9018f3e8f4ec7480a94f57c2397a81c7da51

# Stage1: build
FROM ${GOLANG_IMAGE} as build

COPY TAG /

WORKDIR /go/src/github.com/cilium/certgen
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL "https://github.com/cilium/certgen/archive/v${VERSION}.tar.gz" | \
      tar xzf - --strip-components 1 \
    && CGO_ENABLED=0 go build -o cilium-certgen main.go

# Stage2: runtime
FROM ${BASE_IMAGE}
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/src/github.com/cilium/certgen/cilium-certgen /usr/bin/cilium-certgen
COPY --from=build /go/src/github.com/cilium/certgen/LICENSE  /LICENSE

USER 10000:10000

ENTRYPOINT ["/usr/bin/cilium-certgen"]
