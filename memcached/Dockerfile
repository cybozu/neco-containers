# memcached container

# Stage1: build from sourceFROM quay.io/cybozu/ubuntu-dev:20.04 AS build
FROM quay.io/cybozu/ubuntu-dev:20.04 AS build

ARG MEMCACHED_VERSION=1.6.10

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        automake \
        libevent-dev

RUN git clone --depth=1 -b ${MEMCACHED_VERSION} https://github.com/memcached/memcached /work/memcached

WORKDIR /work/memcached
RUN ./autogen.sh \
    && ./configure --prefix=/usr/local/memcached \
    && make \
    && make install

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
       libevent-dev \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/memcached /usr/local/memcached/
COPY --from=build /work/memcached/LICENSE* /

USER 10000:10000
EXPOSE 11211
ENTRYPOINT ["/usr/local/memcached/bin/memcached"]
