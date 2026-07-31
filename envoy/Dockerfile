# Envoy container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./pkg/probe

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04.20260731@sha256:1b00ea85de4e622400e23054583f3f63f88b559bc04aadc599e124cfb59e0469
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
