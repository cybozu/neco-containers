# BIRD container

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:24.04.20260710@sha256:e779b8941433835b11adf9140f12bb5486168486592063339bacf4cc8ab4ef65 AS build
ENV BIRD_HASH=5b275f0f91c063bd0369a2996982f8f578736eb9b4b5c0fb62a275ec663f74df

ARG BIRD_VERSION=3.2.2
WORKDIR /work

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
       iproute2 \
       flex \
       bison \
       libncurses5-dev \
       libssh-dev \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://bird.nic.cz/download/bird-${BIRD_VERSION}.tar.gz -o bird-${BIRD_VERSION}.tar.gz \
    && echo "${BIRD_HASH} bird-${BIRD_VERSION}.tar.gz" | sha256sum -c - \
    && tar xzf bird-${BIRD_VERSION}.tar.gz

WORKDIR /work/bird-${BIRD_VERSION}

RUN ./configure \
      --prefix=/usr/local/bird \
      --sysconfdir=/etc/bird \
      --with-runtimedir=/run/bird \
      --with-protocols=bgp,rpki,bfd,pipe,static \
    && make \
    && make install

COPY bird-copyright /usr/local/bird/copyright


# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04.20260710@sha256:3f0d95a6fd61317a3eb6ea9e47391aacd7f4129f12a7092e20729e16157ab83b
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
       libssh-4 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/bird /usr/local/bird/
COPY --from=build /etc/bird /etc/bird/
COPY --from=build /etc/iproute2 /etc/iproute2/

ENV PATH=/usr/local/bird/sbin:"$PATH"

ENTRYPOINT ["/usr/local/bird/sbin/bird", "-f"]
