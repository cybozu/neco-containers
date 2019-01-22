# alertmanager container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG ALERTMANAGER_VERSION=0.15.3

RUN git clone https://github.com/prometheus/alertmanager /go/src/github.com/prometheus/alertmanager

WORKDIR /go/src/github.com/prometheus/alertmanager

RUN git checkout tags/v${ALERTMANAGER_VERSION} \
    && GO15VENDOREXPERIMENT=1 go get github.com/prometheus/alertmanager/cmd/...

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu-minimal:bionic-20181112

COPY --from=build /go/bin /usr/local/alertmanager/bin
COPY --from=build /go/src/github.com/prometheus/alertmanager/LICENSE /usr/local/alertmanager/LICENSE

ENV PATH=/usr/local/alertmanager/bin:"$PATH"

RUN mkdir -p /usr/local/alertmanager/bin && chown 10000:10000 /usr/local/alertmanager/bin \
    && mkdir -p /data && chown 10000:10000 /data

USER       10000:10000
EXPOSE     9093
VOLUME     "/data"
ENTRYPOINT [ "alertmanager" ]
