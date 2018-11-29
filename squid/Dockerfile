# squid container

FROM quay.io/cybozu/ubuntu:18.04

ARG SQUID_VERSION=3.5.27-1

RUN apt-get update \
&& apt-get -y install --no-install-recommends \
squid=${SQUID_VERSION}* \
&& rm -rf /var/lib/apt/lists/* \
&& ln -sf /dev/stdout /var/log/squid/access.log \
&& ln -sf /dev/stdout /var/log/squid/store.log \
&& ln -sf /dev/stdout /var/log/squid/cache.log

ADD entrypoint.sh /entrypoint.sh

EXPOSE 3128

ENTRYPOINT ["/entrypoint.sh"]
