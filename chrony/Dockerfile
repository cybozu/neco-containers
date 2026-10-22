# chrony container

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:26.04.20261021@sha256:e8cea4ac6112e219658cda415a875c72d5ae9ed8a6c7df4a070e7777a558bb80 AS build
ENV CHRONY_VERSION=4.9
ENV CHRONY_HASH=4924c6f530105bcd5b9e9e33c48a2ae1bfd889222c8480bc41601110efc864d0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir /work \
    && curl -sSLf https://chrony-project.org/releases/chrony-${CHRONY_VERSION}.tar.gz -o /tmp/chrony.tar.gz \
    && echo "${CHRONY_HASH} /tmp/chrony.tar.gz" | sha256sum -c - \
    && tar zxf /tmp/chrony.tar.gz -C /work/ \
    && apt-get update \
    && apt-get install -y --no-install-recommends libedit-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /work/chrony-${CHRONY_VERSION}
RUN ./configure --exec-prefix=/usr/local/chrony --with-pidfile=/run/chrony/chrony.pid \
    && make \
    && make install \
    && cp ./COPYING /usr/local/chrony/copyright


# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:26.04.20261021@sha256:4d40515de878b3276fa0b0fd6ee4818f7dc729d579d3a211dae0763bfad4ff02
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update \
    && apt-get install -y --no-install-recommends libedit2 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/chrony /usr/local/chrony
COPY chrony.conf.example /etc/chrony.conf

VOLUME /var/lib/chrony/

EXPOSE 123/udp

ENV PATH=/usr/local/chrony/bin:"$PATH"

ENTRYPOINT [ "/usr/local/chrony/sbin/chronyd", "-d"]
