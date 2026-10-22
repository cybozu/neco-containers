FROM ghcr.io/cybozu/ubuntu-dev:26.04.20261021@sha256:e8cea4ac6112e219658cda415a875c72d5ae9ed8a6c7df4a070e7777a558bb80 AS build

ARG UNBOUND_VERSION=1.26.2
ARG UNBOUND_SHA256=d8ee7aa6bbb3ff1be7d9e57637d0d0ec2f18273a8efff1a0afe2c665858b96f7

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && apt-get -y install --no-install-recommends libexpat1-dev libevent-dev
RUN curl -sSLf -o unbound-${UNBOUND_VERSION}.tar.gz https://nlnetlabs.nl/downloads/unbound/unbound-${UNBOUND_VERSION}.tar.gz \
    && echo "${UNBOUND_SHA256}  unbound-${UNBOUND_VERSION}.tar.gz" | sha256sum -c - \
    && tar zxf unbound-${UNBOUND_VERSION}.tar.gz

WORKDIR /work/unbound-${UNBOUND_VERSION}
RUN mkdir -p /usr/local/unbound \
    && ./configure --with-libevent --prefix=/usr/local/unbound \
    && make \
    && make install \
    && mv LICENSE /usr/local/unbound/LICENSE

# It fails on the first time by design.
# To see it works as intended, we run the command twice and check the second exit code.
# See: https://unbound.docs.nlnetlabs.nl/en/latest/manpages/unbound-anchor.html#exit-code
RUN /usr/local/unbound/sbin/unbound-anchor -v || true
RUN /usr/local/unbound/sbin/unbound-anchor -v

FROM ghcr.io/cybozu/ubuntu:26.04.20261021@sha256:4d40515de878b3276fa0b0fd6ee4818f7dc729d579d3a211dae0763bfad4ff02
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update \
    && apt-get install -y --no-install-recommends libevent-2.1-7 libexpat1 \
    && rm -rf /var/lib/apt/lists/*
COPY reload-unbound /usr/local/bin/reload-unbound
COPY --from=build /usr/local/unbound/LICENSE /usr/local/unbound/LICENSE
COPY --from=build /usr/local/unbound/lib /usr/local/unbound/lib
COPY --from=build /usr/local/unbound/sbin /usr/local/unbound/sbin
COPY --from=build /usr/local/unbound/etc/unbound/root.key /usr/local/unbound/etc/unbound/root.key
COPY root.hints /usr/local/unbound/etc/unbound/root.hints

ENV PATH=/usr/local/unbound/sbin:/usr/local/bin:"$PATH"
EXPOSE 53 53/udp
ENTRYPOINT ["unbound", "-d", "-p"]
