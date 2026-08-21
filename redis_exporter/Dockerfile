# redis_exporter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG REDIS_EXPORTER_VERSION=1.89.0

WORKDIR /work/redis_exporter

# ref: https://github.com/oliver006/redis_exporter/blob/025e911c26d8ce7aa23b574909455301fb00dde7/Makefile#L82-L93
RUN curl -fsSL -o redis_exporter.tar.gz "https://github.com/oliver006/redis_exporter/archive/v${REDIS_EXPORTER_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f redis_exporter.tar.gz \
    && rm -f redis_exporter.tar.gz \
    && CGO_ENABLED=0 go build -ldflags="-X main.BuildVersion=v${REDIS_EXPORTER_VERSION} -X main.BuildDate=$(date +%F-%T)" -o redis_exporter .

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/redis_exporter/redis_exporter /redis_exporter
COPY --from=build /work/redis_exporter/LICENSE /LICENSE

USER 10000:10000

EXPOSE 9121

ENTRYPOINT ["/redis_exporter"]
