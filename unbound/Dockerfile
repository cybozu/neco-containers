FROM quay.io/cybozu/ubuntu-dev:18.04 AS build

ARG UNBOUND_VERSION=1.8.1

WORKDIR /work
RUN apt-get update && apt-get -y install --no-install-recommends libexpat1-dev
RUN curl -sSLf https://nlnetlabs.nl/downloads/unbound/unbound-${UNBOUND_VERSION}.tar.gz | \
    tar zxf - -C /work/

WORKDIR /work/unbound-${UNBOUND_VERSION}
RUN mkdir -p /usr/local/unbound \
    && ./configure --prefix=/usr/local/unbound \
    && make \
    && make install \
    && mv LICENSE /usr/local/unbound/LICENSE

FROM quay.io/cybozu/ubuntu:18.04

RUN rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/unbound/LICENSE /usr/local/unbound/LICENSE
COPY --from=build /usr/local/unbound/sbin/unbound /usr/local/unbound/sbin/unbound
COPY --from=build /usr/local/unbound/sbin/unbound-anchor /usr/local/unbound/sbin/unbound-anchor
COPY --from=build /usr/local/unbound/sbin/unbound-checkconf /usr/local/unbound/sbin/unbound-checkconf
COPY --from=build /usr/local/unbound/sbin/unbound-control /usr/local/unbound/sbin/unbound-control
COPY --from=build /usr/local/unbound/sbin/unbound-control-setup /usr/local/unbound/sbin/unbound-control-setup
COPY --from=build /usr/local/unbound/sbin/unbound-host /usr/local/unbound/sbin/unbound-host

ENV PATH=/usr/local/unbound/sbin:"$PATH"
EXPOSE 53 53/udp
ENTRYPOINT ["unbound", "-d"]
