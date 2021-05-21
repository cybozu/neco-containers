# Grafana container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.16-focal AS build
ARG GRAFANA_VERSION=7.5.7

WORKDIR /work/grafana
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/grafana/grafana/archive/v${GRAFANA_VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/grafana

RUN go run build.go build

# Install Node.js 14.x & yarn
# https://github.com/nodesource/distributions/blob/master/README.md#debinstall
RUN curl -sL https://deb.nodesource.com/setup_14.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global yarn

RUN yarn install --pure-lockfile --no-progress
ENV NODE_ENV production
RUN yarn build

# Stage2: setup runtime container
# refer to : https://github.com/grafana/grafana/blob/v7.3.7/Dockerfile#L35-L80
FROM quay.io/cybozu/ubuntu:20.04
ARG GF_UID="10000"
ARG GF_GID="10000"

ENV PATH=/usr/share/grafana/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
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
             "$GF_PATHS_LOGS" \
             "$GF_PATHS_PLUGINS" \
             "$GF_PATHS_DATA" && \
    cp "$GF_PATHS_HOME/conf/sample.ini" "$GF_PATHS_CONFIG" && \
    cp "$GF_PATHS_HOME/conf/ldap.toml" /etc/grafana/ldap.toml && \
    chown -R 10000:10000 "$GF_PATHS_DATA" "$GF_PATHS_HOME/.aws" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING" && \
    chmod 777 "$GF_PATHS_DATA" "$GF_PATHS_HOME/.aws" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING"

COPY --from=build /work/grafana/bin/linux-amd64/grafana-server /work/grafana/bin/linux-amd64/grafana-cli ./bin/
COPY --from=build /work/grafana/public ./public
COPY --from=build /work/grafana/tools ./tools
COPY --from=build /work/grafana/LICENSE ./LICENSE

EXPOSE 3000

COPY --from=build /work/grafana/packaging/docker/run.sh /run.sh

USER 10000:10000
ENTRYPOINT [ "/run.sh" ]
