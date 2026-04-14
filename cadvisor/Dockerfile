# cadvisor container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.25-noble AS build

ARG CADVISOR_COMMIT=7cbdee9260c4292d5d27d00416ce8a7dc1dc12a0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/google/cadvisor
RUN curl -fsSL -o cadvisor.tar.gz "https://github.com/google/cadvisor/archive/${CADVISOR_COMMIT}.tar.gz" \
    && tar -x -z --strip-components 1 -f cadvisor.tar.gz \
    && rm -f cadvisor.tar.gz \
    && cd cmd \
    && CGO_ENABLED=0 go build -tags netgo -ldflags="-w -s" -o cadvisor .

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/src/github.com/google/cadvisor/cmd/cadvisor /cadvisor
COPY --from=build /go/src/github.com/google/cadvisor/LICENSE /LICENSE

EXPOSE 8080

ENTRYPOINT ["/cadvisor", "-logtostderr"]
