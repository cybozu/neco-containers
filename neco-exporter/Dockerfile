FROM ghcr.io/cybozu/golang:1.27.3.2_noble@sha256:d78937fddb26c4350824d4022f5c9dfe20bf762811ea31065f8c6d467c58bc30 AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o neco-exporter

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/neco-exporter /
EXPOSE 8080/tcp
ENTRYPOINT ["/neco-exporter"]
