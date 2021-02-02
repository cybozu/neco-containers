# Prometheus container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.15-focal AS build

ARG PROMETHEUS_VERSION=2.24.1

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install Node 12.x & yarn
# https://github.com/nodesource/distributions/blob/master/README.md#debinstall
RUN curl -sL https://deb.nodesource.com/setup_12.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global yarn

RUN curl -fsSL -o prometheus.tar.gz "https://github.com/prometheus/prometheus/archive/v${PROMETHEUS_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f prometheus.tar.gz \
    && rm -f prometheus.tar.gz \
    && make build npm_licenses

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /work/prometheus /usr/local/prometheus/bin/prometheus
COPY --from=build /work/LICENSE /usr/local/prometheus/LICENSE
COPY --from=build /work/console_libraries /usr/share/prometheus/console_libraries
COPY --from=build /work/consoles /usr/share/prometheus/consoles
# hadolint ignore=DL3010
COPY --from=build /work/npm_licenses.tar.bz2 /usr/local/prometheus/npm_licenses.tar.bz2

RUN mkdir -m 755 /data && chown 10000:10000 /data
VOLUME "/data"
USER 10000:10000

EXPOSE 9090

ENTRYPOINT ["/usr/local/prometheus/bin/prometheus"]
CMD        ["--config.file=/etc/prometheus/prometheus.yaml", \
            "--storage.tsdb.path=/data", \
            "--web.enable-admin-api", \
            "--web.console.libraries=/usr/share/prometheus/console_libraries", \
            "--web.console.templates=/usr/share/prometheus/consoles"]
