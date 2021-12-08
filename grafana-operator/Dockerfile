# Grafana Operator container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.17-focal AS build
ARG VERSION=3.10.3

WORKDIR /work/grafana-operator
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/grafana-operator/grafana-operator/archive/v${VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/grafana-operator

RUN CGO_ENABLED=0 GO111MODULE=on go build -o /usr/local/bin/grafana-operator ./cmd/manager

# https://github.com/grafana-operator/grafana-operator/blob/master/.gitmodules
RUN git clone --depth 1 https://github.com/grafana/grafonnet-lib

# Stage2: setup runtime container
# refer to : https://github.com/grafana-operator/grafana-operator/blob/master/build/Dockerfile
FROM quay.io/cybozu/ubuntu:20.04
USER 10000:10000

COPY --from=build /work/grafana-operator/grafonnet-lib/grafonnet/ /opt/jsonnet/grafonnet
COPY --from=build /work/grafana-operator/grafonnet-lib/LICENSE /opt/jsonnet/LICENSE.grafonnet
COPY --from=build /usr/local/bin/grafana-operator /usr/local/bin/grafana-operator
COPY --from=build /work/grafana-operator/LICENSE /usr/local/grafana-operator/LICENSE
