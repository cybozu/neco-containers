# ref. https://github.com/docker-library/haproxy/blob/master/3.0/Dockerfile
FROM ghcr.io/cybozu/ubuntu-dev:24.04.20260902@sha256:d1d788d7de6d89d8253484b4f6d959a86005aee02ef2b69d9f451fba86be6a4d AS build

ARG HAPROXY_VERSION=3.0.27
ARG HAPROXY_URL=https://www.haproxy.org/download/3.0/src/haproxy-${HAPROXY_VERSION}.tar.gz
ARG HAPROXY_SHA256=c200b72ed5078d99d4bd658834478058409420b45d08dcec7e7a015e9c7b1dd1

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

FROM ghcr.io/cybozu/ubuntu:24.04.20260902@sha256:182a16198fafecbce2f813a07de269d28efb885aa4843e611ce44c25ce46cc2b
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
