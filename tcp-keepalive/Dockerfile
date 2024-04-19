# tcp-keepalive container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build
COPY src /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./tcp-keepalive

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin /
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/tcp-keepalive", "server"]
