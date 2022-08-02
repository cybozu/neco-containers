FROM quay.io/cybozu/ubuntu-dev:20.04 AS build

ARG UNBOUND_VERSION=1.16.1

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && apt-get -y install --no-install-recommends libexpat1-dev
RUN curl -sSLf https://nlnetlabs.nl/downloads/unbound/unbound-${UNBOUND_VERSION}.tar.gz | \
    tar zxf - -C /work/

WORKDIR /work/unbound-${UNBOUND_VERSION}
RUN mkdir -p /usr/local/unbound \
    && ./configure --prefix=/usr/local/unbound \
    && make \
    && make install \
    && mv LICENSE /usr/local/unbound/LICENSE

FROM quay.io/cybozu/ubuntu:20.04

COPY reload-unbound /usr/local/bin/reload-unbound
COPY --from=build /usr/local/unbound/LICENSE /usr/local/unbound/LICENSE
COPY --from=build /usr/local/unbound/sbin /usr/local/unbound/sbin

ENV PATH=/usr/local/unbound/sbin:/usr/local/bin:"$PATH"
EXPOSE 53 53/udp
ENTRYPOINT ["unbound", "-d", "-p"]
