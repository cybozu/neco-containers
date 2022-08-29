FROM quay.io/cybozu/golang:1.18-focal as build

COPY / /work
WORKDIR /work

RUN go build -ldflags="-w -s" .

# stage2: final image
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /work/heartbeat /usr/local/bin/heartbeat
USER 10000:10000

ENTRYPOINT ["/usr/local/bin/heartbeat"]
