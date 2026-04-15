# chrony container

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:24.04 AS build
ENV CHRONY_VERSION=4.8
ENV CHRONY_HASH=33ea8eb2a4daeaa506e8fcafd5d6d89027ed6f2f0609645c6f149b560d301706

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
FROM ghcr.io/cybozu/ubuntu:24.04
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
