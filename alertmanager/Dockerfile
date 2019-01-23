# alertmanager container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG ALERTMANAGER_VERSION=0.15.3

WORKDIR /go/src/github.com/prometheus/alertmanager

RUN curl -fsSL -o alertmanager.tar.gz https://github.com/prometheus/alertmanager/archive/v${ALERTMANAGER_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f alertmanager.tar.gz \
    && rm -f alertmanager.tar.gz \
    && make PREFIX=$GOPATH/bin build \
    && mkdir /data

# Stage2: setup runtime container
FROM scratch

COPY --from=build --chown=10000:10000 /data /data
COPY --from=build --chown=10000:10000 /go/bin /usr/local/alertmanager/bin
COPY --from=build /go/src/github.com/prometheus/alertmanager/LICENSE /usr/local/alertmanager/LICENSE

ENV PATH=/usr/local/alertmanager/bin:"$PATH"

USER       10000:10000
EXPOSE     9093
VOLUME     "/data"
ENTRYPOINT [ "alertmanager" ]
