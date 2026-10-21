FROM ghcr.io/cybozu/golang:1.27.3.1_resolute@sha256:56354d624535147898457543f53b6b42bebd671817e0ec322b610b1370c98f2c AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o squid-exporter

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/squid-exporter /
EXPOSE 8080/tcp
ENTRYPOINT ["/squid-exporter"]
