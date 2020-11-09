# Prometheus container (prometheus, alertmanager, pushgateway)

# Stage1: build from source
FROM quay.io/cybozu/golang:1.15-bionic AS build

ARG PROMETHEUS_VERSION=2.21.0
ARG ALERTMANAGER_VERSION=0.21.0
ARG PUSHGATEWAY_VERSION=1.3.0

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN apt-get update \
    && apt-get install --no-install-recommends -y gnupg \
    && curl -sL https://deb.nodesource.com/setup_10.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
    && echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list \
    && apt-get update && apt-get install --no-install-recommends -y yarn

WORKDIR /go/src/github.com/prometheus/prometheus
RUN curl -fsSL -o prometheus.tar.gz "https://github.com/prometheus/prometheus/archive/v${PROMETHEUS_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f prometheus.tar.gz \
    && rm -f prometheus.tar.gz \
    && make "PREFIX=$GOPATH/bin/prometheus" build

WORKDIR /go/src/github.com/prometheus/alertmanager
RUN curl -fsSL -o alertmanager.tar.gz "https://github.com/prometheus/alertmanager/archive/v${ALERTMANAGER_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f alertmanager.tar.gz \
    && rm -f alertmanager.tar.gz \
    && make "PREFIX=$GOPATH/bin/alertmanager" build

WORKDIR /go/src/github.com/prometheus/pushgateway
RUN curl -fsSL -o pushgateway.tar.gz "https://github.com/prometheus/pushgateway/archive/v${PUSHGATEWAY_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f pushgateway.tar.gz \
    && rm -f pushgateway.tar.gz \
    && make "PREFIX=$GOPATH/bin/pushgateway" build

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

RUN mkdir -m 755 /data && chown 10000:10000 /data
COPY --from=build /go/bin/prometheus /usr/local/prometheus/bin
COPY --from=build /go/bin/alertmanager /usr/local/alertmanager/bin
COPY --from=build /go/bin/pushgateway /usr/local/pushgateway/bin
COPY --from=build /go/src/github.com/prometheus/prometheus/LICENSE /usr/local/prometheus/LICENSE
COPY --from=build /go/src/github.com/prometheus/alertmanager/LICENSE /usr/local/alertmanager/LICENSE
COPY --from=build /go/src/github.com/prometheus/pushgateway/LICENSE /usr/local/pushgateway/LICENSE

ENV PATH=/usr/local/prometheus/bin:/usr/local/alertmanager/bin:/usr/local/pushgateway/bin:"$PATH"

USER 10000:10000

# Prometheus
EXPOSE 9090
# Alertmanager
EXPOSE 9093
# Pushgateway
EXPOSE 9091

VOLUME "/data"
