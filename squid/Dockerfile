# squid container

FROM ghcr.io/cybozu/ubuntu-dev:24.04 AS build

ARG SQUID_VERSION=6.13

WORKDIR /work
# refer https://salsa.debian.org/squid-team/squid/-/blob/master/debian/rules
RUN curl -sfLO https://github.com/squid-cache/squid/releases/download/SQUID_$(echo ${SQUID_VERSION} | sed 's/[.]/_/g')/squid-${SQUID_VERSION}.tar.xz \
    && tar --strip-components=1 -xf /work/squid-${SQUID_VERSION}.tar.xz \
    && CFLAGS="-mavx2" CXXFLAGS="-mavx2" ./configure --without-gnutls --with-openssl --without-systemd \
                   --sysconfdir=/etc/squid --with-swapdir=/var/spool/squid \
                   --with-logdir=/var/log/squid --with-pidfile=/run/squid.pid \
                   --with-filedescriptors=65536 --with-large-files --disable-arch-native \
    && make -j "$(nproc)" \
    && make install

# stage2: production image
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /usr/local/squid /usr/local/squid
COPY --from=build /etc/squid /etc/squid

# Redirect logs to stdout/stderr for the container
RUN mkdir -p /var/log/squid \
    && chown 10000:10000 /var/log/squid \
    && echo 'pid_filename none' >>/etc/squid/squid.conf \
    && echo 'logfile_rotate 0' >>/etc/squid/squid.conf \
    && echo 'access_log stdio:/dev/stdout' >>/etc/squid/squid.conf \
    && echo 'cache_log stdio:/dev/stderr' >>/etc/squid/squid.conf

# Note that the default squid.conf does not enable disk cache.
# /var/spool/squid is only used for coredumps.
RUN mkdir -p /var/spool/squid \
    && chown -R 10000:10000 /var/spool/squid
VOLUME /var/spool/squid

ENV PATH=/usr/local/squid/sbin:/usr/local/squid/bin:$PATH
USER 10000:10000
EXPOSE 3128

ENTRYPOINT ["/usr/local/squid/sbin/squid", "-N"]
