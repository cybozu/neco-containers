# BIRD container

# Stage1: build from source
FROM quay.io/cybozu/ubuntu-dev:18.04 AS build

ARG BIRD_VERSION=2.0.2
WORKDIR /work

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
       iproute2 \
       flex \
       bison \
       libncurses5-dev \
    && rm -rf /var/lib/apt/lists/*

RUN curl -s -O ftp://bird.network.cz/pub/bird/bird-${BIRD_VERSION}.tar.gz && \
    tar xzf bird-${BIRD_VERSION}.tar.gz

WORKDIR bird-${BIRD_VERSION}

RUN ./configure \
      --prefix=/usr/local/bird \
      --sysconfdir=/etc/bird \
      --with-runtimedir=/run/bird \
      --with-protocols=bgp,bfd,static \
    && make \
    && make install

COPY bird-copyright /usr/local/bird/copyright


# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /usr/local/bird /usr/local/bird/
COPY --from=build /etc/bird /etc/bird/
COPY --from=build /etc/iproute2 /etc/iproute2/

ENV PATH=/usr/local/bird/sbin:"$PATH"

ENTRYPOINT ["/usr/local/bird/sbin/bird", "-f"]
