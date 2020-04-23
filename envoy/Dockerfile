# Envoy container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 GO111MODULE=on go install -mod=vendor -ldflags="-w -s" ./pkg/livenessprobe

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY workspace/envoy /usr/local/bin/envoy
COPY workspace/docker-entrypoint.sh /docker-entrypoint.sh
COPY workspace/LICENSE /usr/local/share/doc/envoy/LICENSE
COPY envoy.yaml /etc/envoy/envoy.yaml
COPY --from=build /go/bin /

EXPOSE 9901

USER 10000:10000
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]
