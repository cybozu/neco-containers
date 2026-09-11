# syntax=docker/dockerfile:1
# Grafana container

ARG GRAFANA_VERSION=13.1.0
# Follow the major version of JS_IMAGE (node:24-alpine) in the upstream Dockerfile.
ARG NODE_VERSION=24.20.0

FROM scratch AS grafana-src
ARG GRAFANA_VERSION
ADD --checksum=sha256:3fdc33bce23123b5da795f985df22f047aafe62c4915bb5567fd374e5570418c \
    https://github.com/grafana/grafana/archive/v${GRAFANA_VERSION}.tar.gz /grafana.tar.gz

FROM scratch AS node-download
ARG NODE_VERSION
ADD --checksum=sha256:855d581f8a4eb1a8117e3426de25fe02770592febcfb31369aee1ffbfee9e8ec \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz /node.tar.gz

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build
ARG NODE_VERSION

ENV NODE_OPTIONS=--max_old_space_size=8000

WORKDIR /work/grafana
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN --mount=type=bind,from=grafana-src,source=/grafana.tar.gz,target=/tmp/grafana.tar.gz \
    tar zxf /tmp/grafana.tar.gz --strip-components 1 -C /work/grafana
COPY *.patch /work/
RUN patch -p1 < /work/viewers_can_explore.patch
RUN patch -p1 < /work/disable-unversioned-plugin-install.patch

# Install Node.js (SHA256-pinned tarball, no curl | bash) & yarn
RUN --mount=type=bind,from=node-download,source=/node.tar.gz,target=/tmp/node.tar.gz \
    tar xf /tmp/node.tar.gz -C /opt
ENV PATH="/opt/node-v${NODE_VERSION}-linux-x64/bin:${PATH}"
# `node` container image, which is used to build this product in the upstream, includes yarn 1.22.22 at this time.
# The actual build uses Yarn 4 via yarnPath in the upstream .yarnrc.yml; this classic yarn only bootstraps it.
RUN npm install --global --ignore-scripts yarn@1.22.22

RUN yarn install --immutable
ENV NODE_ENV=production
RUN yarn build

RUN go mod download
RUN make build-go

# Stage2: setup runtime container
# refer to : https://github.com/grafana/grafana/blob/v10.1.5/Dockerfile#L76-L165
FROM ghcr.io/cybozu/ubuntu:24.04
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
             "$GF_PATHS_HOME/data/plugins-bundled" \
             "$GF_PATHS_DATA" && \
    cp "conf/sample.ini" "$GF_PATHS_CONFIG" && \
    cp "conf/ldap.toml" /etc/grafana/ldap.toml && \
    chown -R 10000:10000 "$GF_PATHS_DATA" "$GF_PATHS_HOME/.aws" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING" && \
    chmod -R 777 "$GF_PATHS_DATA" "$GF_PATHS_HOME/.aws" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING"

COPY --from=build /work/grafana/bin/grafana* /work/grafana/bin/*/grafana* ./bin/
COPY --from=build /work/grafana/public ./public
COPY --from=build /work/grafana/LICENSE ./LICENSE
COPY --from=build /work/grafana/conf/defaults.ini ./conf/defaults.ini

EXPOSE 3000

COPY --from=build /work/grafana/packaging/docker/run.sh /run.sh

USER 10000:10000
ENTRYPOINT [ "/run.sh" ]
