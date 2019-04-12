FROM quay.io/cybozu/ubuntu-dev:18.04 as build

RUN apt-get update && \
    apt-get install -y --no-install-recommends wget && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV GOSU_VERSION 1.10
RUN dpkgArch="$(dpkg --print-architecture | awk -F- '{ print $NF }')" && \
    wget -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$dpkgArch" && \
    chmod +x /usr/local/bin/gosu && \
    gosu nobody true

ENV REDIS_VERSION 5.0.4
ENV REDIS_DOWNLOAD_URL http://download.redis.io/releases/redis-5.0.4.tar.gz

RUN wget -O redis.tar.gz "$REDIS_DOWNLOAD_URL" && \
    mkdir -p /usr/src/redis && \
    tar -xzf redis.tar.gz -C /usr/src/redis --strip-components=1 && \
    rm redis.tar.gz && \
    grep -q '^#define CONFIG_DEFAULT_PROTECTED_MODE 1$' /usr/src/redis/src/server.h && \
    sed -ri 's!^(#define CONFIG_DEFAULT_PROTECTED_MODE) 1$!\1 0!' /usr/src/redis/src/server.h && \
    grep -q '^#define CONFIG_DEFAULT_PROTECTED_MODE 0$' /usr/src/redis/src/server.h && \
    make -C /usr/src/redis -j "$(nproc)" && \
    make -C /usr/src/redis install

FROM quay.io/cybozu/ubuntu:18.04

RUN groupadd -g 10000 -r redis && \
    useradd -r -u 10000 -g redis redis && \
    mkdir /data && chown redis:redis /data && \
    mkdir -p /usr/local/redis

VOLUME /data
WORKDIR /data

COPY --from=build /usr/local/bin/* /usr/local/redis/bin/
COPY --from=build /usr/src/redis/COPYING /usr/local/redis/COPYING

ENV PATH=/usr/local/redis/bin:/usr/local/bin:"$PATH"

EXPOSE 6379
USER 10000:10000

ENTRYPOINT ["redis-server"]
