# Grafana container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build
ARG GRAFANA_VERSION=6.5.2

WORKDIR /work/grafana
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/grafana/grafana/archive/v${GRAFANA_VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/grafana

RUN GOFLAGS=-mod=vendor make build-go

RUN apt-get update && \
    apt-get install -y --no-install-recommends gnupg && \
    curl -sL https://deb.nodesource.com/setup_10.x | bash - && \
    apt-get install -y --no-install-recommends nodejs

RUN npm install -g yarn && \
    yarn install --pure-lockfile --no-progress
ENV NODE_ENV production
RUN ./node_modules/.bin/grunt build

# Stage2: setup runtime container
# refer to : https://github.com/grafana/grafana/blob/653918056c594d7f56a65771b2c9681bdf8a3b9a/Dockerfile#L37-L83
FROM quay.io/cybozu/ubuntu:18.04
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

RUN apt-get update && \
    apt-get install --no-install-recommends -qq -y \
    libfontconfig && \
    rm -rf /var/lib/apt/lists/*
COPY --from=build /work/grafana/conf ./conf

RUN groupadd -r -g $GF_GID grafana && \
    useradd -r -u $GF_UID -g grafana grafana && \
    mkdir -p "$GF_PATHS_PROVISIONING/datasources" \
             "$GF_PATHS_PROVISIONING/dashboards" \
             "$GF_PATHS_LOGS" \
             "$GF_PATHS_PLUGINS" \
             "$GF_PATHS_DATA" && \
    cp "$GF_PATHS_HOME/conf/sample.ini" "$GF_PATHS_CONFIG" && \
    cp "$GF_PATHS_HOME/conf/ldap.toml" /etc/grafana/ldap.toml && \
    chown -R 10000:10000 "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" && \
    chmod 755 "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS"

COPY --from=build /work/grafana/bin/linux-amd64/grafana-server /work/grafana/bin/linux-amd64/grafana-cli ./bin/
COPY --from=build /work/grafana/public ./public
COPY --from=build /work/grafana/tools ./tools

EXPOSE 3000

COPY --from=build /work/grafana/packaging/docker/run.sh /run.sh

USER 10000:10000
ENTRYPOINT [ "/run.sh" ]
