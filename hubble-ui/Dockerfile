ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.22-jammy
ARG UBUNTU_IMAGE=ghcr.io/cybozu/ubuntu:22.04
ARG BACKEND_IMAGE=scratch
ARG NGINX_VERSION=1.23.4
ARG NJS_VERSION=0.7.11
ARG PKG_RELEASE=1~jammy
ARG NGINX_UNPRIVILEGED_COMMIT_HASH=582fa68a7609f2d3dad7fb2121af0244b3ce976c

# Stage 1: build
FROM ${GOLANG_IMAGE} AS builder-base
COPY TAG /

WORKDIR /app

RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL "https://github.com/cilium/hubble-ui/archive/v${VERSION}.tar.gz" | \
      tar xzf - --strip-components 1

# Stage 1: build hubble-ui
FROM builder-base as build-hubble-ui
ARG NODE_VERSION=18

WORKDIR /app

RUN curl -sSLf https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install
ARG NODE_ENV=production
RUN npm run build

# Stage 1: build hubble-ui-backend
FROM builder-base as build-hubble-ui-backend
WORKDIR /app/backend
# hubble-ui-backend
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o backend \
    && apt-get update \
    && apt-get install -y --no-install-recommends binutils-aarch64-linux-gnu \
    && ./build-gops.sh

# Stage 2: hubble-ui runtime
FROM ${UBUNTU_IMAGE} as frontend
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG NGINX_VERSION
ARG NJS_VERSION
ARG PKG_RELEASE
ARG NGINX_UNPRIVILEGED_COMMIT_HASH

ARG UID=101
ARG GID=101

RUN set -x \
    # create nginx user/group first, to be consistent throughout docker variants
    && addgroup --system --gid $GID nginx || true \
    && adduser --system --disabled-login --ingroup nginx --no-create-home --home /nonexistent --gecos "nginx user" --shell /bin/false --uid $UID nginx || true \
    && apt-get update \
    && apt-get install --no-install-recommends --no-install-suggests -y gnupg1 ca-certificates \
    && \
    NGINX_GPGKEY=573BFD6B3D8FBC641079A6ABABF5BD827BD9BF62; \
    found=''; \
    for server in \
        hkp://keyserver.ubuntu.com:80 \
        pgp.mit.edu \
    ; do \
        echo "Fetching GPG key $NGINX_GPGKEY from $server"; \
        apt-key adv --keyserver "$server" --keyserver-options timeout=10 --recv-keys "$NGINX_GPGKEY" && found=yes && break; \
    done; \
    test -z "$found" && echo >&2 "error: failed to fetch GPG key $NGINX_GPGKEY" && exit 1; \
    apt-get remove --purge --auto-remove -y gnupg1 && rm -rf /var/lib/apt/lists/* \
    && nginxPackages=" \
        nginx=${NGINX_VERSION}-${PKG_RELEASE} \
        nginx-module-xslt=${NGINX_VERSION}-${PKG_RELEASE} \
        nginx-module-geoip=${NGINX_VERSION}-${PKG_RELEASE} \
        nginx-module-image-filter=${NGINX_VERSION}-${PKG_RELEASE} \
        nginx-module-njs=${NGINX_VERSION}+${NJS_VERSION}-${PKG_RELEASE} \
    " \
    && echo "deb https://nginx.org/packages/mainline/ubuntu/ jammy nginx" >> /etc/apt/sources.list.d/nginx.list \
    && apt-get update \
    && apt-get install --no-install-recommends --no-install-suggests -y \
                        $nginxPackages \
                        gettext-base \
                        curl \
    && apt-get remove --purge --auto-remove -y && rm -rf /var/lib/apt/lists/* /etc/apt/sources.list.d/nginx.list \
    \
# if we have leftovers from building, let's purge them (including extra, unnecessary build deps)
    && if [ -n "$tempDir" ]; then \
        apt-get purge -y --auto-remove \
        && rm -rf "$tempDir" /etc/apt/sources.list.d/temp.list; \
    fi \
# forward request and error logs to docker log collector
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
# create a docker-entrypoint.d directory
    && mkdir /docker-entrypoint.d \
    && for entrypoint in \
            20-envsubst-on-templates.sh \
            30-tune-worker-processes.sh \
        ; do \
            curl -o /docker-entrypoint.d/${entrypoint} -fsSL https://raw.githubusercontent.com/nginxinc/docker-nginx-unprivileged/${NGINX_UNPRIVILEGED_COMMIT_HASH}/mainline/debian/${entrypoint}; \
    done \
    && curl -o /docker-entrypoint.sh -fsSL https://raw.githubusercontent.com/nginxinc/docker-nginx-unprivileged/${NGINX_UNPRIVILEGED_COMMIT_HASH}/mainline/debian/docker-entrypoint.sh \
    && chmod +x /docker-entrypoint.d/* \
    && chmod +x /docker-entrypoint.sh

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

COPY --from=build-hubble-ui /app/server/public /app

ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080

STOPSIGNAL SIGQUIT

USER $UID

CMD ["nginx", "-g", "daemon off;"]

# Stage 2: hubble-ui-backend runtime
FROM ${BACKEND_IMAGE} as backend
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-hubble-ui-backend /app/backend/backend /
COPY --from=build-hubble-ui-backend /out/linux/amd64/bin/gops /usr/bin/gops

ENTRYPOINT ["/backend"]
