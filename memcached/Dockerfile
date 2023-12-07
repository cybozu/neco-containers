# memcached container

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:22.04 AS build

ARG MEMCACHED_VERSION=1.6.22

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        automake \
        libevent-dev \
        pkg-config

RUN git clone --depth=1 -b ${MEMCACHED_VERSION} https://github.com/memcached/memcached /work/memcached

WORKDIR /work/memcached
RUN ./autogen.sh \
    && ./configure --prefix=/usr/local/memcached \
    && make \
    && make install

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
       libevent-dev \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/memcached /usr/local/memcached/
COPY --from=build /work/memcached/LICENSE* /

USER 10000:10000
EXPOSE 11211
ENTRYPOINT ["/usr/local/memcached/bin/memcached"]
