# victoriametrics-operator container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG SRCREPO=VictoriaMetrics/victoriametrics-datasource
ARG VERSION=0.19.7

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install Node.js 20.x and yarn
# https://github.com/nodesource/distributions/blob/master/README.md#debinstall
RUN curl -sSLf https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global yarn@1.22.22
    # `node` container image, which is used to build this product in the upsteram, includes yarn 1.22.22 at this time.

WORKDIR /go/src/github.com/VictoriaMetrics/victoriametrics-datasource
RUN curl -fsSL -o victoriametrics-datasource.tar.gz "https://github.com/${SRCREPO}/archive/v${VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f victoriametrics-datasource.tar.gz \
    && rm -f victoriametrics-datasource.tar.gz
COPY yarn-install-immutable.patch /yarn-install-immutable.patch
RUN patch -p1 < /yarn-install-immutable.patch
RUN make vm-backend-plugin-build
RUN yarn preinstall && yarn install --immutable --omit=dev && yarn build && yarn sign --distDir plugins/victoriametrics-metrics-datasource

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/src/github.com/VictoriaMetrics/victoriametrics-datasource/plugins/victoriametrics-metrics-datasource /plugins/victoriametrics-metrics-datasource
COPY --from=build /go/src/github.com/VictoriaMetrics/victoriametrics-datasource/LICENSE /LICENSE

USER 10000:10000

ENV GRAFANA_PLUGINS_DIR=/var/lib/grafana/plugins

ENTRYPOINT ["bash", "-c"]
CMD ["mkdir -p ${GRAFANA_PLUGINS_DIR} && cp -r /plugins/victoriametrics-metrics-datasource ${GRAFANA_PLUGINS_DIR}/"]
