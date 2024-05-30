# Build the controller binary
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY internal/controller/ internal/controller/

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o sts-updater cmd/main.go

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /workspace/sts-updater .
USER 10000:10000
ENTRYPOINT ["/sts-updater"]
