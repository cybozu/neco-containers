# BIRD container

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:26.04.20261021@sha256:e8cea4ac6112e219658cda415a875c72d5ae9ed8a6c7df4a070e7777a558bb80 AS build
ENV BIRD_HASH=8042eb36a76388400143ef36d4734b28572c4ad2c3f35c5dbdd2936acc2fff2c

ARG BIRD_VERSION=3.2.4
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
FROM ghcr.io/cybozu/ubuntu:26.04.20261021@sha256:4d40515de878b3276fa0b0fd6ee4818f7dc729d579d3a211dae0763bfad4ff02
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
       libssh-4 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/bird /usr/local/bird/
COPY --from=build /etc/bird /etc/bird/
# iproute2 >= 6.x (Ubuntu 26.04) installs its config files under /usr/share/iproute2
COPY --from=build /usr/share/iproute2 /etc/iproute2/

ENV PATH=/usr/local/bird/sbin:"$PATH"

ENTRYPOINT ["/usr/local/bird/sbin/bird", "-f"]
