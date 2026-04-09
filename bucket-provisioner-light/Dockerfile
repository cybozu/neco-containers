FROM ghcr.io/cybozu/golang:1.26-noble AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o /out/bucket-provisioner-light ./main.go

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /out/bucket-provisioner-light /bucket-provisioner-light
USER 10000:10000

ENTRYPOINT ["/bucket-provisioner-light"]
