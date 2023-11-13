FROM ghcr.io/cybozu/ubuntu:20.04 AS build

ENV HAPROXY_BRANCH 2.6
ENV HAPROXY_MINOR 2.6.14
ENV HAPROXY_SHA256 bd3dd9fa60391ca09e1225e1ac3163e45be83c3f54f2fd76a30af289cc6e4fd4
ENV HAPROXY_SRC_URL http://www.haproxy.org/download
ENV DEBIAN_FRONTEND noninteractive

# if you use ubuntu:22.04, you must use "libssl3=3.0.2-0ubuntu1.9". but HA-Proxy 2.6 does NOT adapt it.
# We choose "ubuntu:20.04" and "libssl1.1" to get a stable operation.

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    procps \
    zlib1g \
    libssl1.1 \
    "libpcre2-*" \
    liblua5.3-0 \
    libatomic1 \
    tar \
    curl \
    socat \
    gcc \
    make \
    libc6-dev \
    libssl-dev \
    libpcre3-dev \
    zlib1g-dev \
    liblua5.3-dev && \
    curl -sfSL "${HAPROXY_SRC_URL}/${HAPROXY_BRANCH}/src/haproxy-${HAPROXY_MINOR}.tar.gz" -o haproxy.tar.gz && \
    echo "$HAPROXY_SHA256 *haproxy.tar.gz" | sha256sum -c - && \
    mkdir -p /tmp/haproxy && \
    tar -xzf haproxy.tar.gz -C /tmp/haproxy --strip-components=1 && \
    rm -f haproxy.tar.gz && \
    make -C /tmp/haproxy -j"$(nproc)" TARGET=linux-glibc CPU=generic USE_PCRE2=1 USE_PCRE2_JIT=1 USE_OPENSSL=1 \
                            USE_TFO=1 USE_LINUX_TPROXY=1 USE_LUA=1 USE_GETADDRINFO=1 \
                            USE_PROMEX=1 USE_SLZ=1 \
                            all && \
    make -C /tmp/haproxy TARGET=linux-glibc install-bin install-man


FROM ghcr.io/cybozu/ubuntu:20.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers/haproxy"

ENV HAPROXY_UID haproxy
ENV HAPROXY_GID haproxy
ENV DEBIAN_FRONTEND noninteractive

RUN mkdir -p /var/lib/haproxy && \
    mkdir -p /usr/local/haproxy && \
    mkdir -p /usr/local/bin

COPY --from=build /usr/local/sbin/haproxy /usr/local/bin/haproxy
COPY --from=build /tmp/haproxy/examples/errorfiles /usr/etc/haproxy/errors
COPY --from=build /tmp/haproxy/LICENSE /usr/local/haproxy/LICENSE
COPY docker-entrypoint.sh /

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    procps \
    zlib1g \
    libssl1.1 \
    "libpcre2-*" \
    liblua5.3-0 \
    libatomic1 && \
    groupadd "$HAPROXY_GID" && \
    useradd -g "$HAPROXY_GID" "$HAPROXY_UID" && \
    chown "$HAPROXY_UID:$HAPROXY_GID" /var/lib/haproxy && \
    mkdir -p /usr/local/etc/haproxy && \
    ln -s /usr/local/etc/haproxy /etc/haproxy && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

STOPSIGNAL SIGUSR1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["haproxy", "-f", "/usr/local/etc/haproxy/haproxy.cfg"]
