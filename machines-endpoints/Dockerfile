# machines-endpoints container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.4.2_resolute@sha256:0ddbebb97d9fefa0961dd10fdccedc40a4b805d28f21f21d92faa2f7b071d2f1 AS build
ENV GOPROXY=https://golang.flatt.tech

COPY . /work

WORKDIR /work

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    CGO_ENABLED=0 go install -ldflags="-w -s" ./pkg/machines-endpoints

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin /

USER 10000:10000

ENTRYPOINT ["/machines-endpoints"]
