# Build stage
FROM ghcr.io/cybozu/golang:1.27.3.2_noble@sha256:d78937fddb26c4350824d4022f5c9dfe20bf762811ea31065f8c6d467c58bc30 AS builder

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
