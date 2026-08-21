# redis_exporter container

FROM scratch AS redis-exporter-src
ARG REDIS_EXPORTER_VERSION=1.89.0
ADD --checksum=sha256:fdc98abb490a21f32dcb4020ef662401279aedfa2daf4aca983e84570ce48acc \
    https://github.com/oliver006/redis_exporter/archive/v${REDIS_EXPORTER_VERSION}.tar.gz /src.tar.gz

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.7.1_noble@sha256:678dffc420bde6361513f3c63a5c9018f3e8f4ec7480a94f57c2397a81c7da51 AS build

ARG REDIS_EXPORTER_VERSION

WORKDIR /work

COPY --from=redis-exporter-src /src.tar.gz /src.tar.gz

# ref: https://github.com/oliver006/redis_exporter/blob/025e911c26d8ce7aa23b574909455301fb00dde7/Dockerfile#L20-L21
# Diffs:
#   - `-ldflags "-s -w"` is removed to keep debug information for troubleshooting.
#   - `-extldflags "-static"` is removed since it is not necessary when CGO_ENABLED=0.
#   - `-X main.BuildCommitSha=$SHA1` is removed since it is not necessary for our use case.
#   - `-trimpath` is added to remove all file system paths from the compiled executable to make it more reproducible.
RUN tar -x -z --strip-components 1 -f /src.tar.gz \
 && BUILD_DATE=$(date +%F-%T) CGO_ENABLED=0 go build -trimpath -o redis_exporter \
    -ldflags="-X main.BuildVersion=v${REDIS_EXPORTER_VERSION} -X main.BuildDate=$BUILD_DATE"  .

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/redis_exporter /redis_exporter
COPY --from=build /work/LICENSE /LICENSE

USER 10000:10000

EXPOSE 9121

ENTRYPOINT ["/redis_exporter"]
