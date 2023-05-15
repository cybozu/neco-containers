#FROM ubuntu:jammy-20230425
FROM ubuntu:focal-20230412

MAINTAINER Maho Takara <takara@cybozu.co.jp>

LABEL Name HAProxy
LABEL Release Community Edition
LABEL Vendor HAProxy
LABEL Version 2.6.10

ENV HAPROXY_BRANCH 2.6
ENV HAPROXY_MINOR 2.6.10
ENV HAPROXY_SHA256 e71b2cd9ca1043345f083a5225078ccf824dced2b5779d86f11fa4e88f451773
ENV HAPROXY_SRC_URL http://www.haproxy.org/download

ENV HAPROXY_UID haproxy
ENV HAPROXY_GID haproxy

ENV DEBIAN_FRONTEND noninteractive


# if you use ubuntu:jammy, you must use "libssl3=3.0.2-0ubuntu1.9". but it is few case.
# We choose "ubuntu:focal" and "libssl1.1" to get a stable operation.

RUN apt-get update && \
    apt-get install -y --no-install-recommends procps zlib1g libssl1.1 "libpcre2-*" liblua5.3-0 libatomic1 tar curl socat ca-certificates && \
    apt-get install -y --no-install-recommends gcc make libc6-dev libssl-dev libpcre3-dev zlib1g-dev liblua5.3-dev && \
    curl -sfSL "${HAPROXY_SRC_URL}/${HAPROXY_BRANCH}/src/haproxy-${HAPROXY_MINOR}.tar.gz" -o haproxy.tar.gz && \
    echo "$HAPROXY_SHA256 *haproxy.tar.gz" | sha256sum -c - && \
    groupadd "$HAPROXY_GID" && \
    useradd -g "$HAPROXY_GID" "$HAPROXY_UID" && \
    mkdir -p /tmp/haproxy && \
    tar -xzf haproxy.tar.gz -C /tmp/haproxy --strip-components=1 && \
    rm -f haproxy.tar.gz && \
    make -C /tmp/haproxy -j"$(nproc)" TARGET=linux-glibc CPU=generic USE_PCRE2=1 USE_PCRE2_JIT=1 USE_OPENSSL=1 \
                            USE_TFO=1 USE_LINUX_TPROXY=1 USE_LUA=1 USE_GETADDRINFO=1 \
                            USE_PROMEX=1 USE_SLZ=1 \
                            all && \
    make -C /tmp/haproxy TARGET=linux-glibc install-bin install-man # && \
    ln -s /usr/local/sbin/haproxy /usr/sbin/haproxy && \
    mkdir -p /var/lib/haproxy && \
    chown "$HAPROXY_UID:$HAPROXY_GID" /var/lib/haproxy && \
    mkdir -p /usr/local/etc/haproxy && \
    ln -s /usr/local/etc/haproxy /etc/haproxy && \
    cp -R /tmp/haproxy/examples/errorfiles /usr/local/etc/haproxy/errors && \
    rm -rf /tmp/haproxy && \
    apt-get purge -y --auto-remove gcc make libc6-dev libssl-dev libpcre2-dev zlib1g-dev liblua5.3-dev && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /

STOPSIGNAL SIGUSR1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["haproxy", "-f", "/usr/local/etc/haproxy/haproxy.cfg"]
