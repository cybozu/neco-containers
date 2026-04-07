# Refer to this Dockerfile: https://github.com/redis/docker-library-redis/blob/v8.2.5/debian/Dockerfile

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:24.04 AS build

ARG REDIS_VERSION=8.2.5
ARG REDIS_DOWNLOAD_URL=https://github.com/redis/redis/archive/refs/tags/${REDIS_VERSION}.tar.gz
ARG REDIS_DOWNLOAD_SHA=1ddb3c3de137ee67a533be769b79294f036328fe6995caff1006462a4342641f

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# necessary packages (ca-certificates dpkg-dev gcc g++ libc6-dev libssl-dev make wget)
# are pre-installed except for wget. wget can be substituted with curl
RUN curl -sSLf "$REDIS_DOWNLOAD_URL" -o redis.tar.gz \
    && echo "$REDIS_DOWNLOAD_SHA *redis.tar.gz" | sha256sum -c -
RUN mkdir -p /usr/src/redis \
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
    && extraJemallocConfigureFlags="$extraJemallocConfigureFlags --with-lg-page=12" \
    && extraJemallocConfigureFlags="$extraJemallocConfigureFlags --with-lg-hugepage=21" \
    && sed -ri 's!cd jemalloc && ./configure !&'"$extraJemallocConfigureFlags"' !' /usr/src/redis/deps/Makefile \
    \
    && BUILD_TLS=yes make -C /usr/src/redis PREFIX=/usr/local/redis -j "$(nproc)" all \
    && BUILD_TLS=yes make -C /usr/src/redis PREFIX=/usr/local/redis install

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN groupadd -g 10000 -r redis && \
    useradd -r -u 10000 -g redis redis && \
    mkdir /data && chown redis:redis /data

VOLUME /data
WORKDIR /data

COPY --from=build /usr/local/redis /usr/local/redis
COPY --from=build /usr/src/redis/LICENSE.txt /usr/local/redis/LICENSE.txt
COPY --from=build /usr/src/redis/deps/fpconv/LICENSE.txt /usr/local/redis/licenses/LICENSE-fpconv
COPY --from=build /usr/src/redis/deps/hdr_histogram/COPYING.txt /usr/local/redis/licenses/COPYING-hdr_histogram
COPY --from=build /usr/src/redis/deps/hdr_histogram/LICENSE.txt /usr/local/redis/licenses/LICENSE-hdr_histogram
COPY --from=build /usr/src/redis/deps/hiredis/COPYING /usr/local/redis/licenses/COPYING-hiredis
COPY --from=build /usr/src/redis/deps/jemalloc/COPYING /usr/local/redis/licenses/COPYING-jemalloc
COPY --from=build /usr/src/redis/deps/lua/COPYRIGHT /usr/local/redis/licenses/LICENSE-lua

ENV PATH=/usr/local/redis/bin:"$PATH"

EXPOSE 6379
USER 10000:10000

# the line below is different from upstream Dockerfile
# if you want to use this outside of Neco, you may need to edit ENTRYPOINT and COMMAND appropriately
ENTRYPOINT ["redis-server"]
