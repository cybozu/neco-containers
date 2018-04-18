# lldpd container

# Stage1: build from source
FROM quay.io/cybozu/ubuntu-dev:18.04 AS build

ARG LLDPD_VERSION=1.0.1
WORKDIR /work

RUN curl -s -O https://media.luffy.cx/files/lldpd/lldpd-${LLDPD_VERSION}.tar.gz \
    && tar -x -z -f lldpd-${LLDPD_VERSION}.tar.gz

WORKDIR lldpd-${LLDPD_VERSION}

RUN ./configure \
      --prefix=/usr/local/lldpd \
      --sysconfdir=/etc/lldpd \
      --runstatedir=/run/lldpd \
      --disable-static \
      --with-privsep-user=root \
      --with-privsep-group=root \
    && make \
    && make install

RUN cp libevent/LICENSE /usr/local/lldpd/share/doc/lldpd/LICENSE_libevent


# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /usr/local/lldpd /usr/local/lldpd/
COPY --from=build /etc/lldpd /etc/lldpd/

RUN rm /usr/local/lldpd/lib/liblldpctl.so /usr/local/lldpd/lib/liblldpctl.so.4 \
    && echo "/usr/local/lldpd/lib" > /etc/ld.so.conf.d/lldpd.conf \
    && ldconfig

ENV PATH=/usr/local/lldpd/sbin:"$PATH"

ENTRYPOINT ["/usr/local/lldpd/sbin/lldpd", "-d"]
