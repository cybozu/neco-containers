# squid container

FROM quay.io/cybozu/ubuntu:18.04

ARG SQUID_VERSION=3.5.27-1ubuntu1.5

RUN apt-get update \
&& apt-get -y install --no-install-recommends \
squid=${SQUID_VERSION}* \
&& rm -rf /var/lib/apt/lists/*

# Redirect logs to stdout for the container
RUN ln -sf /dev/stdout /var/log/squid/access.log
RUN ln -sf /dev/stdout /var/log/squid/store.log
RUN ln -sf /dev/stdout /var/log/squid/cache.log

COPY entrypoint.sh /entrypoint.sh

EXPOSE 3128
VOLUME /var/spool/squid

ENTRYPOINT ["/entrypoint.sh"]
