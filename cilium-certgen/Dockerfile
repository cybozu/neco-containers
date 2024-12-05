ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.23-jammy

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
