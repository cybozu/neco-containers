# Stage1
FROM quay.io/cybozu/golang:1.20-jammy AS builder

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY main.go main.go
COPY pkg/ pkg/

# Build
RUN pwd
RUN ls /work
RUN CGO_ENABLED=0 go install -ldflags="-w -s"

# Stage2
FROM scratch
COPY --from=builder /go/bin /
USER 10000:10000
ENTRYPOINT ["/necosen"]
