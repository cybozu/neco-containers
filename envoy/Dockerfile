# Envoy container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./pkg/probe

# Stage2: setup runtime container
# Upstream uses the `nonroot` tag (UID 65532) for its distroless image.
# https://github.com/envoyproxy/envoy/blob/ef2d997c1b022cf8b849a1d3521fbf234d79ca26/distribution/docker/Dockerfile-envoy#L62
# We use the `latest` tag instead and run Envoy as the non-privileged `nobody` user
# (UID 65534) via `USER nobody` below, keeping the same UID as the previous image.
FROM gcr.io/distroless/base-nossl-debian12:latest@sha256:e35893f8ca1cabfbd60394af94899c20930c4615355adf97c5da660a51b61bca
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY workspace/envoy /usr/local/bin/envoy
COPY workspace/LICENSE /usr/local/share/doc/envoy/LICENSE
COPY envoy.yaml /etc/envoy/envoy.yaml
COPY --from=build /go/bin/probe /usr/local/bin/probe

EXPOSE 9901

USER nobody
ENTRYPOINT ["/usr/local/bin/envoy"]
CMD ["-c", "/etc/envoy/envoy.yaml"]
