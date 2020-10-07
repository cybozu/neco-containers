# Grafana Operator container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build
ARG VERSION=3.5.0-support-env-from-secret

WORKDIR /work/grafana-operator
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/cybozu-go/grafana-operator/archive/v${VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/grafana-operator

RUN CGO_ENABLED=0 GO111MODULE=on go build -o /usr/local/bin/grafana-operator -mod=vendor ./cmd/manager

# Stage2: setup runtime container
# refer to : https://github.com/cybozu-go/grafana-operator/blob/master/build/Dockerfile
FROM quay.io/cybozu/ubuntu:18.04
USER 10000:10000

COPY --from=build /usr/local/bin/grafana-operator /usr/local/bin/grafana-operator
COPY --from=build /work/grafana-operator/LICENSE /usr/local/grafana-operator/LICENSE
