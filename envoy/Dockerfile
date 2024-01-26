# Envoy container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.20-jammy AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./pkg/probe

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY workspace/envoy /usr/local/bin/envoy
COPY workspace/docker-entrypoint.sh /docker-entrypoint.sh
COPY workspace/LICENSE /usr/local/share/doc/envoy/LICENSE
COPY envoy.yaml /etc/envoy/envoy.yaml
COPY --from=build /go/bin/probe /usr/local/bin/probe

EXPOSE 9901

USER nobody
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]
