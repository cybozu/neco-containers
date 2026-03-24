# Build stage
FROM ghcr.io/cybozu/golang:1.26-noble AS builder

WORKDIR /work
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /client ./cmd/client

# Server image
FROM scratch AS server
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /server /server
ENTRYPOINT ["/server"]

# Client image
FROM scratch AS client
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /client /client
ENTRYPOINT ["/client"]
