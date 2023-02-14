# moco-mysql container

FROM quay.io/cybozu/ubuntu:22.04

COPY workspace /usr/local/mysql/

RUN apt-get update \
  && apt-get install -y --no-install-recommends libjemalloc2 libnuma1 libaio1 \
  && rm -rf /var/lib/apt/lists/* \
  && mkdir -p /var/lib/mysql \
  && chown -R 10000:10000 /var/lib/mysql

ENV PATH=/usr/local/mysql/bin:"$PATH"
VOLUME /var/lib/mysql
ENTRYPOINT ["mysqld"]
EXPOSE 3306 33060 33062 8080
USER 10000:10000
