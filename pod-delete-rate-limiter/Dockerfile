# pod-deletion-rate-limiter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.23-noble AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/pod-delete-rate-limiter cmd/main.go

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/bin/pod-delete-rate-limiter /pod-delete-rate-limiter

USER 10000:10000

ENTRYPOINT ["/pod-delete-rate-limiter"]
