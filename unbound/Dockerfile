FROM ghcr.io/cybozu/ubuntu-dev:24.04.20260917@sha256:721ea828b11e92594f8a7905c4ca2fb7471e54f8fe41c8fe978b8513c10811e1 AS build

ARG UNBOUND_VERSION=1.26.1
ARG UNBOUND_SHA256=35a6dc0e425a9282c3426d9a3043144011bf0534aed4b73ab62c52aee0af1503

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

FROM ghcr.io/cybozu/ubuntu:24.04.20260917@sha256:b2c77f62764d21184562ff35714ecf1dfa22f311c4b3b9cd4d4f46a63e5fd117
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
