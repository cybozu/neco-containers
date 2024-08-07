# Redis container
# Refer to this Dockerfile: https://github.com/redis/docker-library-redis/blob/master/7.0/debian/Dockerfile

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:22.04 as build

ARG GOSU_VERSION=1.17
ARG REDIS_VERSION=7.0.15
ARG REDIS_DOWNLOAD_URL=http://download.redis.io/releases/redis-${REDIS_VERSION}.tar.gz

RUN apt-get update \
    && apt-get install -y --no-install-recommends wget pkg-config \
    && rm -rf /var/lib/apt/lists/* \
    && wget -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/${GOSU_VERSION}/gosu-amd64" \
    && chmod +x /usr/local/bin/gosu \
    && gosu nobody true

RUN wget -O redis.tar.gz "$REDIS_DOWNLOAD_URL" \
    && mkdir -p /usr/src/redis \
    && tar -xzf redis.tar.gz -C /usr/src/redis --strip-components=1 \
# disable Redis protected mode [1] as it is unnecessary in context of Docker
# (ports are not automatically exposed when running inside Docker, but rather explicitly by specifying -p / -P)
# [1]: https://github.com/redis/redis/commit/edd4d555df57dc84265fdfb4ef59a4678832f6da
    && sed -ri 's!^( *createBoolConfig[(]"protected-mode",.*, *)1( *,.*[)],)$!\10\2!' /usr/src/redis/src/config.c \
# for future reference, we modify this directly in the source instead of just supplying a default configuration flag because apparently "if you specify any argument to redis-server, [it assumes] you are going to specify everything"
# see also https://github.com/docker-library/redis/issues/4#issuecomment-50780840
# (more exactly, this makes sure the default behavior of "save on SIGTERM" stays functional by default)
    \
# https://github.com/jemalloc/jemalloc/issues/467 -- we need to patch the "./configure" for the bundled jemalloc to match how Debian compiles, for compatibility
# (also, we do cross-builds, so we need to embed the appropriate "--build=xxx" values to that "./configure" invocation)
    && extraJemallocConfigureFlags="--build=x86_64-linux-gnu" \
# https://salsa.debian.org/debian/jemalloc/-/blob/c0a88c37a551be7d12e4863435365c9a6a51525f/debian/rules#L8-23
    && extraJemallocConfigureFlags="$extraJemallocConfigureFlags --with-lg-hugepage=21" \
    && sed -ri 's!cd jemalloc && ./configure !&'"$extraJemallocConfigureFlags"' !' /usr/src/redis/deps/Makefile \
    \
    && BUILD_TLS=yes make -C /usr/src/redis PREFIX=/usr/local/redis -j "$(nproc)" all \
    && BUILD_TLS=yes make -C /usr/src/redis PREFIX=/usr/local/redis install

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN groupadd -g 10000 -r redis && \
    useradd -r -u 10000 -g redis redis && \
    mkdir /data && chown redis:redis /data

VOLUME /data
WORKDIR /data

COPY --from=build /usr/local/redis /usr/local/redis
COPY --from=build /usr/src/redis/COPYING /usr/local/redis/COPYING

ENV PATH=/usr/local/redis/bin:"$PATH"

EXPOSE 6379
USER 10000:10000

# the line below is different from upstream Dockerfile
# if you want to use this outside of Neco, you may need to edit ENTRYPOINT and COMMAND apporpriately
ENTRYPOINT ["redis-server"]
