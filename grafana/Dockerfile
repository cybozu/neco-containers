# Grafana container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build
ARG GRAFANA_VERSION=5.4.3

WORKDIR /go/src/github.com/grafana/grafana
RUN curl -sSLf https://github.com/grafana/grafana/archive/v${GRAFANA_VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /go/src/github.com/grafana/grafana

RUN make build-go

RUN apt-get update && \
    apt-get install -y gnupg && \
    curl -sL https://deb.nodesource.com/setup_8.x | bash - && \
    apt-get install -y nodejs

RUN npm install -g yarn && \
    yarn install --pure-lockfile --no-progress
ENV NODE_ENV production
RUN ./node_modules/.bin/grunt build

# Stage2: setup runtime container
# refer to : https://github.com/grafana/grafana/blob/653918056c594d7f56a65771b2c9681bdf8a3b9a/Dockerfile#L37-L83
FROM quay.io/cybozu/ubuntu:18.04
ARG GF_UID="472"
ARG GF_GID="472"

ENV PATH=/usr/share/grafana/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    GF_PATHS_CONFIG="/etc/grafana/grafana.ini" \
    GF_PATHS_DATA="/var/lib/grafana" \
    GF_PATHS_HOME="/usr/share/grafana" \
    GF_PATHS_LOGS="/var/log/grafana" \
    GF_PATHS_PLUGINS="/var/lib/grafana/plugins" \
    GF_PATHS_PROVISIONING="/etc/grafana/provisioning"

WORKDIR $GF_PATHS_HOME

RUN apt-get update && apt-get upgrade -y && \
    apt-get install -qq -y \
    libfontconfig \
    ca-certificates && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /go/src/github.com/grafana/grafana/conf ./conf

RUN groupadd -r -g $GF_GID grafana && \
    useradd -r -u $GF_UID -g grafana grafana && \
    mkdir -p "$GF_PATHS_PROVISIONING/datasources" \
             "$GF_PATHS_PROVISIONING/dashboards" \
             "$GF_PATHS_LOGS" \
             "$GF_PATHS_PLUGINS" \
             "$GF_PATHS_DATA" && \
    cp "$GF_PATHS_HOME/conf/sample.ini" "$GF_PATHS_CONFIG" && \
    cp "$GF_PATHS_HOME/conf/ldap.toml" /etc/grafana/ldap.toml && \
    chown -R grafana:grafana "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" && \
    chmod 777 "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS"

COPY --from=build /go/src/github.com/grafana/grafana/bin/linux-amd64/grafana-server /go/src/github.com/grafana/grafana/bin/linux-amd64/grafana-cli ./bin/
COPY --from=build /go/src/github.com/grafana/grafana/public ./public
COPY --from=build /go/src/github.com/grafana/grafana/tools ./tools

EXPOSE 3000

COPY --from=build /go/src/github.com/grafana/grafana/packaging/docker/run.sh /run.sh

USER grafana
ENTRYPOINT [ "/run.sh" ]
