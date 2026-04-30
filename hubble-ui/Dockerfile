ARG NODE_VERSION=22.22.2
ARG NODE_SHA=88fd1ce767091fd8d4a99fdb2356e98c819f93f3b1f8663853a2dee9b438068a
ARG NGINX_VERSION=1.29.1
ARG NJS_VERSION=0.9.1
ARG PKG_RELEASE=1~noble

# ==================== backend-builder image ====================
# https://github.com/cilium/hubble-ui/blob/master/backend/Dockerfile
FROM ghcr.io/cybozu/golang:1.25-noble AS backend-builder

WORKDIR /app
COPY src/hubble-ui .

WORKDIR /app/backend
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o backend

# ==================== hubble-ui-backend image ====================
FROM scratch AS backend
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"
COPY --from=backend-builder /app/backend/backend /
ENTRYPOINT ["/backend"]

# ==================== frontend-builder image ====================
# https://github.com/cilium/hubble-ui/blob/master/Dockerfile
FROM ghcr.io/cybozu/golang:1.25-noble AS frontend-builder
ARG NODE_VERSION
ARG NODE_SHA

WORKDIR /app
COPY src/hubble-ui .
COPY pnpm-lock.yaml .
# pnpm binary is fetched and SHA256-verified on the host by `make pnpm-bin`.
COPY --chmod=0755 .pnpm-bin/pnpm /usr/local/bin/pnpm

# SHA256-pinned Node.js (avoids unauthenticated curl|bash from nodesource).
RUN curl -fsSL -O https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz \
    && echo "${NODE_SHA}  node-v${NODE_VERSION}-linux-x64.tar.xz" | sha256sum -c - \
    && tar xJf node-v${NODE_VERSION}-linux-x64.tar.xz -C /opt \
    && mv /opt/node-v${NODE_VERSION}-linux-x64 /opt/node \
    && rm node-v${NODE_VERSION}-linux-x64.tar.xz
ENV PATH="/opt/node/bin:${PATH}"

# Sequence:
#   1. frozen install of upstream deps (lockfile covers upstream only)
#   2. sass dedupe via pnpm.overrides (instanceof fails with duplicate copies)
#   3. phantom deps upstream's webpack imports but doesn't declare
#      (registry fetch; pinned to pkg@version)
#   4. root postinstall (patch-package + ./scripts/post-install.sh)
# --ignore-scripts blocks lifecycle scripts; root postinstall is explicit.
RUN pnpm install --frozen-lockfile --ignore-scripts \
    && pnpm pkg set "pnpm.overrides.sass=1.89.2" \
    && pnpm install -D --ignore-scripts \
        @protobuf-ts/runtime@2.11.1 \
        @protobuf-ts/runtime-rpc@2.11.1 \
        @types/node@24.1.0 \
        mobx-react-lite@4.1.0 \
    && pnpm run postinstall

ARG NODE_ENV=production
RUN pnpm run build

# ==================== hubble-ui-frontend image ====================
# https://github.com/nginx/docker-nginx-unprivileged/blob/main/Dockerfile-debian.template
FROM ghcr.io/cybozu/ubuntu:24.04 AS frontend
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"
ARG NGINX_VERSION
ARG NJS_VERSION
ARG PKG_RELEASE
ARG UID=101
ARG GID=101

RUN set -x \
# create nginx user/group first, to be consistent throughout docker variants
    && groupadd --system --gid $GID nginx || true \
    && useradd --system --gid nginx --no-create-home --home /nonexistent --comment "nginx user" --shell /bin/false --uid $UID nginx || true \
    && apt-get update \
    && apt-get install --no-install-recommends --no-install-suggests -y gnupg1 ca-certificates \
    && \
    NGINX_GPGKEYS="573BFD6B3D8FBC641079A6ABABF5BD827BD9BF62 8540A6F18833A80E9C1653A42FD21310B49F6B46 9E9BE90EACBCDE69FE9B204CBCDCD8A38D88A2B3"; \
    NGINX_GPGKEY_PATH=/etc/apt/keyrings/nginx-archive-keyring.gpg; \
    export GNUPGHOME="$(mktemp -d)"; \
    found=''; \
    for NGINX_GPGKEY in $NGINX_GPGKEYS; do \
        for server in \
            hkp://keyserver.ubuntu.com:80 \
            pgp.mit.edu \
        ; do \
            echo "Fetching GPG key $NGINX_GPGKEY from $server"; \
            gpg1 --batch --keyserver "$server" --keyserver-options timeout=10 --recv-keys "$NGINX_GPGKEY" && found=yes && break; \
        done; \
        test -z "$found" && echo >&2 "error: failed to fetch GPG key $NGINX_GPGKEY" && exit 1; \
    done; \
    gpg1 --batch --export $NGINX_GPGKEYS > "$NGINX_GPGKEY_PATH" ; \
    rm -rf "$GNUPGHOME"; \
    apt-get remove --purge --auto-remove -y gnupg1 && rm -rf /var/lib/apt/lists/* \
    && dpkgArch="$(dpkg --print-architecture)" \
# check src/docker-nginx-unprivileged/update.sh
    && nginxPackages=" \
        nginx=${NGINX_VERSION}-${PKG_RELEASE} \
        nginx-module-xslt=${NGINX_VERSION}-${PKG_RELEASE} \
        nginx-module-geoip=${NGINX_VERSION}-${PKG_RELEASE} \
        nginx-module-image-filter=${NGINX_VERSION}-${PKG_RELEASE} \
        nginx-module-njs=${NGINX_VERSION}+${NJS_VERSION}-${PKG_RELEASE} \
    " \
    && case "$dpkgArch" in \
        amd64|arm64) \
# arches officialy built by upstream
            echo "deb [signed-by=$NGINX_GPGKEY_PATH] https://nginx.org/packages/mainline/ubuntu/ noble nginx" >> /etc/apt/sources.list.d/nginx.list \
            && apt-get update \
            ;; \
        *) \
# we don't need to handle architectures other than amd64/arm64
            exit 1;; \
    esac \
    \
    && apt-get install --no-install-recommends --no-install-suggests -y \
                        $nginxPackages \
                        gettext-base \
                        curl \
    && apt-get remove --purge --auto-remove -y && rm -rf /var/lib/apt/lists/* /etc/apt/sources.list.d/nginx.list \
    \
# skip tempDir treatment, because it was for non-amd64/arm64
# forward request and error logs to docker log collector
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
# create a docker-entrypoint.d directory
    && mkdir /docker-entrypoint.d

# implement changes required to run NGINX as an unprivileged user
RUN sed -i 's,listen       80;,listen       8080;,' /etc/nginx/conf.d/default.conf \
    && sed -i '/user  nginx;/d' /etc/nginx/nginx.conf \
    && sed -i 's,\(/var\)\{0\,1\}/run/nginx.pid,/tmp/nginx.pid,' /etc/nginx/nginx.conf \
    && sed -i "/^http {/a \    proxy_temp_path /tmp/proxy_temp;\n    client_body_temp_path /tmp/client_temp;\n    fastcgi_temp_path /tmp/fastcgi_temp;\n    uwsgi_temp_path /tmp/uwsgi_temp;\n    scgi_temp_path /tmp/scgi_temp;\n" /etc/nginx/nginx.conf \
    && sed -i 's,PIDFILE=${PIDFILE:-/run/nginx.pid},PIDFILE=${PIDFILE:-/tmp/nginx.pid},' /etc/init.d/nginx \
# nginx user must own the cache and etc directory to write cache and tweak the nginx config
    && chown -R $UID:0 /var/cache/nginx \
    && chmod -R g+w /var/cache/nginx \
    && chown -R $UID:0 /etc/nginx \
    && chmod -R g+w /etc/nginx

COPY src/docker-nginx-unprivileged/mainline/debian/docker-entrypoint.sh /
COPY src/docker-nginx-unprivileged/mainline/debian/10-listen-on-ipv6-by-default.sh /docker-entrypoint.d
COPY src/docker-nginx-unprivileged/mainline/debian/15-local-resolvers.envsh /docker-entrypoint.d
COPY src/docker-nginx-unprivileged/mainline/debian/20-envsubst-on-templates.sh /docker-entrypoint.d
COPY src/docker-nginx-unprivileged/mainline/debian/30-tune-worker-processes.sh /docker-entrypoint.d

# https://github.com/cybozu/neco-containers/blob/main/hubble-ui/Dockerfile
COPY --from=frontend-builder /app/server/public /app

ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080

STOPSIGNAL SIGQUIT

USER $UID

CMD ["nginx", "-g", "daemon off;"]
