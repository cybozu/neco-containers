# Grafana container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build
ARG GRAFANA_VERSION=10.4.5

ENV NODE_OPTIONS=--max_old_space_size=8000

WORKDIR /work/grafana
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/grafana/grafana/archive/v${GRAFANA_VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/grafana

# Install Node.js 20.x & yarn
# https://github.com/nodesource/distributions/blob/master/README.md#installation-instructions
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global yarn

RUN yarn install --immutable
ENV NODE_ENV production
RUN yarn build

RUN go mod verify
RUN make build-go

# Stage2: setup runtime container
# refer to : https://github.com/grafana/grafana/blob/v10.1.5/Dockerfile#L76-L165
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG GF_UID="10000"
ARG GF_GID="10000"

ENV PATH=/usr/share/grafana/bin:$PATH \
    GF_PATHS_CONFIG="/etc/grafana/grafana.ini" \
    GF_PATHS_DATA="/var/lib/grafana" \
    GF_PATHS_HOME="/usr/share/grafana" \
    GF_PATHS_LOGS="/var/log/grafana" \
    GF_PATHS_PLUGINS="/var/lib/grafana/plugins" \
    GF_PATHS_PROVISIONING="/etc/grafana/provisioning"

WORKDIR $GF_PATHS_HOME

COPY --from=build /work/grafana/conf ./conf

RUN mkdir -p "$GF_PATHS_HOME/.aws" && \
    groupadd -r -g $GF_GID grafana && \
    useradd -r -u $GF_UID -g grafana grafana && \
    mkdir -p "$GF_PATHS_PROVISIONING/datasources" \
             "$GF_PATHS_PROVISIONING/dashboards" \
             "$GF_PATHS_PROVISIONING/notifiers" \
             "$GF_PATHS_PROVISIONING/plugins" \
             "$GF_PATHS_PROVISIONING/access-control" \
             "$GF_PATHS_PROVISIONING/alerting" \
             "$GF_PATHS_LOGS" \
             "$GF_PATHS_PLUGINS" \
             "$GF_PATHS_DATA" && \
    cp "conf/sample.ini" "$GF_PATHS_CONFIG" && \
    cp "conf/ldap.toml" /etc/grafana/ldap.toml && \
    chown -R 10000:10000 "$GF_PATHS_DATA" "$GF_PATHS_HOME/.aws" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING" && \
    chmod -R 777 "$GF_PATHS_DATA" "$GF_PATHS_HOME/.aws" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING"

COPY --from=build /work/grafana/bin/grafana* /work/grafana/bin/*/grafana* ./bin/
COPY --from=build /work/grafana/public ./public
COPY --from=build /work/grafana/LICENSE ./LICENSE

EXPOSE 3000

COPY --from=build /work/grafana/packaging/docker/run.sh /run.sh

USER 10000:10000
ENTRYPOINT [ "/run.sh" ]
