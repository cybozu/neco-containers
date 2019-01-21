# alertmanager container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG ALERTMANAGER_VERSION=0.15.3

WORKDIR /work

RUN curl -fsSL -o alertmanager.tar.gz https://github.com/prometheus/alertmanager/archive/v${ALERTMANAGER_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f alertmanager.tar.gz \
    && rm -f alertmanager.tar.gz \
    && GO15VENDOREXPERIMENT=1 go get github.com/prometheus/alertmanager/cmd/...

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin /usr/local/alertmanager/bin
COPY --from=build /work/LICENSE /usr/local/alertmanager/LICENSE
RUN chown -R nobody:nogroup /usr/local/alertmanager

ENV PATH=/usr/local/alertmanager/bin:"$PATH"

USER       nobody
EXPOSE     9093
VOLUME     "/data"
ENTRYPOINT [ "alertmanager" ]
