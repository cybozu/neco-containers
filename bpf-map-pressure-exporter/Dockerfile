# bpf-map-pressure-exporter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o bpf-map-pressure-exporter

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/bpf-map-pressure-exporter /
EXPOSE 8080/tcp
ENTRYPOINT ["/bpf-map-pressure-exporter"]
