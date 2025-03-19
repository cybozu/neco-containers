# Alertmanager container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG ALERTMANAGER_VERSION=0.28.1

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/prometheus/alertmanager

# Install Node.js 22.x
# https://github.com/nodesource/distributions/blob/master/README.md#installation-instructions
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* 

RUN curl -fsSL -o alertmanager.tar.gz "https://github.com/prometheus/alertmanager/archive/v${ALERTMANAGER_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f alertmanager.tar.gz \
    && rm -f alertmanager.tar.gz \
    && make "PREFIX=$GOPATH/bin/alertmanager" build

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/alertmanager /bin
COPY --from=build /go/src/github.com/prometheus/alertmanager/LICENSE /
COPY --from=build /go/src/github.com/prometheus/alertmanager/COPYRIGHT.txt /
COPY --from=build /go/src/github.com/prometheus/alertmanager/NOTICE /

USER 10000:10000
EXPOSE 9093
ENTRYPOINT ["/bin/alertmanager"]
