# Prometheus container (prometheus, alertmanager, pushgateway)

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG PROMETHEUS_VERSION=2.7.1
ARG ALERTMANAGER_VERSION=0.15.3
ARG PUSHGATEWAY_VERSION=0.7.0

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root

WORKDIR /go/src/github.com/prometheus/prometheus
RUN curl -fsSL -o prometheus.tar.gz https://github.com/prometheus/prometheus/archive/v${PROMETHEUS_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f prometheus.tar.gz \
    && rm -f prometheus.tar.gz \
    && make PREFIX=$GOPATH/bin/prometheus build

WORKDIR /go/src/github.com/prometheus/alertmanager
RUN mkdir /data
RUN curl -fsSL -o alertmanager.tar.gz https://github.com/prometheus/alertmanager/archive/v${ALERTMANAGER_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f alertmanager.tar.gz \
    && rm -f alertmanager.tar.gz \
    && make PREFIX=$GOPATH/bin/alertmanager build

WORKDIR /go/src/github.com/prometheus/pushgateway
RUN curl -fsSL -o pushgateway.tar.gz https://github.com/prometheus/pushgateway/archive/v${PUSHGATEWAY_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f pushgateway.tar.gz \
    && rm -f pushgateway.tar.gz \
    && make PREFIX=$GOPATH/bin/pushgateway build

# Stage2: setup runtime container
FROM scratch

COPY --from=build --chown=10000:10000 /data /data
COPY --from=build --chown=10000:10000 /go/bin/prometheus /usr/local/prometheus/bin
COPY --from=build --chown=10000:10000 /go/bin/alertmanager /usr/local/alertmanager/bin
COPY --from=build --chown=10000:10000 /go/bin/pushgateway /usr/local/pushgateway/bin
COPY --from=build /go/src/github.com/prometheus/prometheus/LICENSE /usr/local/prometheus/LICENSE
COPY --from=build /go/src/github.com/prometheus/alertmanager/LICENSE /usr/local/alertmanager/LICENSE
COPY --from=build /go/src/github.com/prometheus/pushgateway/LICENSE /usr/local/pushgateway/LICENSE

ENV PATH=/usr/local/prometheus/bin:"$PATH"
ENV PATH=/usr/local/alertmanager/bin:"$PATH"
ENV PATH=/usr/local/pushgateway/bin:"$PATH"

USER 10000:10000

# Prometheus
EXPOSE 9090
# Alertmanager
EXPOSE 9093
# Pushgateway
EXPOSE 9091

VOLUME "/data"
