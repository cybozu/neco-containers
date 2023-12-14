FROM ghcr.io/cybozu/golang:1.21-jammy AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -o squid-exporter

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/squid-exporter /
EXPOSE 8080/tcp
ENTRYPOINT ["/squid-exporter"]
