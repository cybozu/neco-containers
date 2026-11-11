# cadvisor container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27-resolute AS build
ENV GOPROXY=https://golang.flatt.tech

ARG CADVISOR_COMMIT=5bf5d43ac6f60d7ee36a13b4d5b4a78ad2d0abc7

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/google/cadvisor
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    curl -fsSL -o cadvisor.tar.gz "https://github.com/google/cadvisor/archive/${CADVISOR_COMMIT}.tar.gz" \
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
