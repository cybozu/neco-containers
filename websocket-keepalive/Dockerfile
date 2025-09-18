FROM ghcr.io/cybozu/golang:1.24-noble AS build

COPY . /work
WORKDIR /work

RUN CGO_ENABLED=0 go build -o websocket-keepalive .

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/websocket-keepalive /usr/local/bin/websocket-keepalive

USER 10000:10000
EXPOSE 9000

ENTRYPOINT ["/usr/local/bin/websocket-keepalive"]
