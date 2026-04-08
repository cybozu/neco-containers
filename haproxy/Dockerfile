# ref. https://github.com/docker-library/haproxy/blob/master/3.0/Dockerfile
FROM ghcr.io/cybozu/ubuntu-dev:24.04 AS build

ARG HAPROXY_VERSION=3.0.19
ARG HAPROXY_URL=https://www.haproxy.org/download/3.0/src/haproxy-${HAPROXY_VERSION}.tar.gz
ARG HAPROXY_SHA256=85f2743a44e4226a0176926fbc8eebfa5551d22fe5f44366877570b6e18c6664

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        gcc \
        libc6-dev \
        liblua5.4-dev \
        libpcre2-dev \
        libssl-dev \
        wget

RUN wget -O haproxy.tar.gz "$HAPROXY_URL" && \
    echo "$HAPROXY_SHA256 *haproxy.tar.gz" | sha256sum -c && \
    mkdir -p /tmp/haproxy && \
    tar -xzf haproxy.tar.gz -C /tmp/haproxy --strip-components=1 && \
    rm haproxy.tar.gz && \
    makeOpts=' \
        TARGET=linux-glibc \
        USE_GETADDRINFO=1 \
        USE_LUA=1 LUA_INC=/usr/include/lua5.4 \
        USE_OPENSSL=1 \
        USE_PCRE2=1 USE_PCRE2_JIT=1 \
        USE_PROMEX=1 \
        USE_PTHREAD_EMULATION=1 \
        ' && \
    make -C /tmp/haproxy -j "$(nproc)" all $makeOpts && \
    make -C /tmp/haproxy install-bin $makeOpts

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

# runtime does not need -dev packages
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        liblua5.4-0 \
        socat && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 99 --system haproxy && \
    useradd \
        --gid haproxy \
        --home-dir /var/lib/haproxy \
        --no-create-home \
        --system \
        --uid 99 \
        haproxy && \
    mkdir -p /usr/local/etc/haproxy && \
    mkdir -p /var/lib/haproxy && \
    chown haproxy:haproxy /var/lib/haproxy

COPY --from=build /usr/local/sbin/haproxy /usr/local/sbin/haproxy
COPY --from=build /tmp/haproxy/examples/errorfiles /usr/local/etc/haproxy/errors
COPY --from=build /tmp/haproxy/LICENSE /usr/local/etc/haproxy/LICENSE

# smoke test
RUN haproxy -v

STOPSIGNAL SIGUSR1

COPY docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]

USER haproxy
WORKDIR /var/lib/haproxy

CMD ["haproxy", "-f", "/usr/local/etc/haproxy/haproxy.cfg"]
