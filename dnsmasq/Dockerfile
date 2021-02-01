# dnsmasq container

# stage1: build
FROM quay.io/cybozu/ubuntu-dev:20.04 as build

ARG DNSMASQ_VERSION=2.84

WORKDIR /work
RUN curl -sfLO http://www.thekelleys.org.uk/dnsmasq/dnsmasq-${DNSMASQ_VERSION}.tar.gz \
    && tar xzf dnsmasq-${DNSMASQ_VERSION}.tar.gz \
    && make -C dnsmasq-${DNSMASQ_VERSION} PREFIX=/usr/local/dnsmasq install \
    && cp dnsmasq-${DNSMASQ_VERSION}/COPYING /usr/local/dnsmasq \
    && cp dnsmasq-${DNSMASQ_VERSION}/COPYING-v3 /usr/local/dnsmasq

# stage2: runtime
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /usr/local/dnsmasq /usr/local/dnsmasq

EXPOSE 67/udp

VOLUME /var/lib/misc

ENTRYPOINT ["/usr/local/dnsmasq/sbin/dnsmasq"]
