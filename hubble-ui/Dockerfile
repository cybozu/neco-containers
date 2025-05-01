ARG NODE_VERSION=22
ARG NGINX_VERSION=1.27.3
ARG NJS_VERSION=0.8.7
ARG PKG_RELEASE=1~noble

# ==================== backend-builder image ====================
# https://github.com/cilium/hubble-ui/blob/master/backend/Dockerfile
FROM ghcr.io/cybozu/golang:1.24-noble AS backend-builder

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
FROM ghcr.io/cybozu/golang:1.24-noble AS frontend-builder
ARG NODE_VERSION

WORKDIR /app
COPY src/hubble-ui .

# https://deb.nodesource.com/
RUN curl -sSLf https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install

ARG NODE_ENV=production
RUN npm run build

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
            gpg1 --keyserver "$server" --keyserver-options timeout=10 --recv-keys "$NGINX_GPGKEY" && found=yes && break; \
        done; \
        test -z "$found" && echo >&2 "error: failed to fetch GPG key $NGINX_GPGKEY" && exit 1; \
    done; \
    gpg1 --export "$NGINX_GPGKEYS" > "$NGINX_GPGKEY_PATH" ; \
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
    && sed -i 's,/var/run/nginx.pid,/tmp/nginx.pid,' /etc/nginx/nginx.conf \
    && sed -i "/^http {/a \    proxy_temp_path /tmp/proxy_temp;\n    client_body_temp_path /tmp/client_temp;\n    fastcgi_temp_path /tmp/fastcgi_temp;\n    uwsgi_temp_path /tmp/uwsgi_temp;\n    scgi_temp_path /tmp/scgi_temp;\n" /etc/nginx/nginx.conf \
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
