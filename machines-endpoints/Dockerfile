# machines-endpoints container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./pkg/machines-endpoints

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin /

USER 10000:10000
EXPOSE 7373 7946

ENTRYPOINT ["/machines-endpoints"]
