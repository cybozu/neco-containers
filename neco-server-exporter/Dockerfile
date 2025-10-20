FROM ghcr.io/cybozu/golang:1.24-noble AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o neco-server-exporter

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/neco-server-exporter /
EXPOSE 8080/tcp
ENTRYPOINT ["/neco-server-exporter"]
