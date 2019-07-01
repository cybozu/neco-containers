# chrony container

# Stage1: build from source
FROM quay.io/cybozu/ubuntu-dev:18.04 AS build
ARG CHRONY_VERSION=3.5

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir /work \
    && curl -sSLf https://download.tuxfamily.org/chrony/chrony-${CHRONY_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /work/chrony-${CHRONY_VERSION}
RUN ./configure --exec-prefix=/usr/local/chrony --with-pidfile=/run/chrony/chrony.pid \
    && make \
    && make install \
    && cp ./COPYING /usr/local/chrony/copyright


# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /usr/local/chrony /usr/local/chrony
COPY chrony.conf.example /etc/chrony.conf

VOLUME /var/lib/chrony/

EXPOSE 123/udp

ENV PATH=/usr/local/chrony/bin:"$PATH"

ENTRYPOINT [ "/usr/local/chrony/sbin/chronyd", "-d"]
