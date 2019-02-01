# Node exporter container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build
ARG NODE_EXPORTER_VERSION=0.17.0

WORKDIR /go/src/github.com/prometheus/node_exporter
RUN curl -sSLf https://github.com/prometheus/node_exporter/archive/v${NODE_EXPORTER_VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /go/src/github.com/prometheus/node_exporter

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root

RUN make PREFIX=${GOPATH}/bin build

# Stage2: setup runtime container
FROM scratch
COPY --from=build --chown=10000:10000 /go/bin /usr/local/node_exporter/bin
COPY --from=build /go/src/github.com/prometheus/node_exporter/LICENSE /usr/local/node_exporter/LICENSE

ENV PATH=/usr/local/node_exporter/bin:"$PATH"

USER 10000:10000
EXPOSE 9100
ENTRYPOINT [ "node_exporter" ]
