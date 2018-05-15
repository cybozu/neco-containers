# dnsmasq container

FROM quay.io/cybozu/ubuntu:18.04

ARG DNSMASQ_VERSION=2.79-1

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
       dnsmasq-base=${DNSMASQ_VERSION} \
    && rm -rf /var/lib/apt/lists/*

EXPOSE 67/udp

VOLUME /var/lib/misc

ENTRYPOINT ["/usr/sbin/dnsmasq"]
