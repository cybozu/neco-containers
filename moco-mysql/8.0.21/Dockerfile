# moco-mysql container

FROM quay.io/cybozu/ubuntu:18.04

ARG MOCO_VERSION=0.3.0

COPY workspace /usr/local/mysql/
COPY ping.sh /ping.sh

RUN apt-get update \
  && apt-get install -y --no-install-recommends libjemalloc1 libnuma1 \
  && rm -rf /var/lib/apt/lists/* \
  && curl -sSL -o /entrypoint https://github.com/cybozu-go/moco/releases/download/v${MOCO_VERSION}/agent \
  && chmod +x /entrypoint \
  && mkdir -p /var/lib/mysql \
  && chown -R 10000:10000 /var/lib/mysql

ENV PATH=/usr/local/mysql/bin:"$PATH"
VOLUME /var/lib/mysql
ENTRYPOINT ["mysqld"]
HEALTHCHECK CMD /ping.sh
EXPOSE 3306 33060 33062 8080
USER 10000:10000
