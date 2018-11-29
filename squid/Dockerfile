# squid container

FROM quay.io/cybozu/ubuntu:18.04

ARG SQUID_VERSION=3.5.27-1

RUN apt-get update \
&& apt-get -y install --no-install-recommends \
squid=${SQUID_VERSION}* \
&& rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /entrypoint.sh

EXPOSE 3128
VOLUME /var/spool/squid
VOLUME /var/log/squid

ENTRYPOINT ["/entrypoint.sh"]
