FROM ghcr.io/cybozu/ubuntu-dev:22.04 AS build

ENV HAPROXY_BRANCH=2.6
ENV HAPROXY_MINOR=2.6.17
ENV HAPROXY_SRC_URL="https://www.haproxy.org/download/${HAPROXY_BRANCH}/src/haproxy-${HAPROXY_MINOR}.tar.gz"
ENV HAPROXY_SHA256=be48ee8ff9127c402b4c6cf1445cef7052f2c540ed1eff2dd04af677b8cd9dd0

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        liblua5.3-dev \
        libpcre2-dev \
        wget && \
    rm -rf /var/lib/apt/lists/* && \
    wget -O haproxy.tar.gz "$HAPROXY_SRC_URL" && \
    echo "$HAPROXY_SHA256 *haproxy.tar.gz" | sha256sum -c && \
    mkdir -p /tmp/haproxy && \
    tar -xzf haproxy.tar.gz -C /tmp/haproxy --strip-components=1 && \
    rm -f haproxy.tar.gz && \
    makeOpts=' \
        TARGET=linux-glibc \
        USE_GETADDRINFO=1 \
        USE_LUA=1 LUA_INC=/usr/include/lua5.3 \
        USE_OPENSSL=1 \
        USE_PCRE2=1 USE_PCRE2_JIT=1 \
        USE_PROMEX=1 \
        ' && \
    make -C /tmp/haproxy -j "$(nproc)" all $makeOpts && \
    make -C /tmp/haproxy install-bin $makeOpts


FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        liblua5.3-dev && \
    rm -rf /var/lib/apt/lists/* && \
    groupadd --gid 99 --system haproxy && \
    useradd \
        --gid haproxy \
        --home-dir /var/lib/haproxy \
        --no-create-home \
        --system \
        --uid 99 \
        haproxy && \
    mkdir -p /usr/local/haproxy && \
    mkdir -p /usr/local/etc/haproxy && \
    mkdir -p /var/lib/haproxy && \
    chown haproxy:haproxy /var/lib/haproxy

COPY --from=build /usr/local/sbin/haproxy /usr/local/bin/haproxy
COPY --from=build /tmp/haproxy/examples/errorfiles /usr/local/etc/haproxy/errors
COPY --from=build /tmp/haproxy/LICENSE /usr/local/haproxy/LICENSE
COPY docker-entrypoint.sh /usr/local/bin/

# smoke test
RUN haproxy -v

STOPSIGNAL SIGUSR1

ENTRYPOINT ["docker-entrypoint.sh"]

USER haproxy

CMD ["haproxy", "-f", "/usr/local/etc/haproxy/haproxy.cfg"]
