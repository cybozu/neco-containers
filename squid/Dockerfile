# squid container

FROM quay.io/cybozu/ubuntu-dev:20.04 as build

ARG SQUID_VERSION=5.2

WORKDIR /work
# refer https://salsa.debian.org/squid-team/squid/-/blob/master/debian/rules
RUN curl -sfLO http://www.squid-cache.org/Versions/v5/squid-${SQUID_VERSION}.tar.xz \
    && tar --strip-components=1 -xf /work/squid-${SQUID_VERSION}.tar.xz \
    && ./configure --without-gnutls --with-openssl --without-systemd \
                   --sysconfdir=/etc/squid --with-swapdir=/var/spool/squid \
                   --with-logdir=/var/log/squid --with-pidfile=/run/squid.pid \
                   --with-filedescriptors=65536 --with-large-files \
    && make -j "$(nproc)" \
    && make install

# stage2: production image
FROM quay.io/cybozu/ubuntu:20.04

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
