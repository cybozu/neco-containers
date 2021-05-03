FROM quay.io/cybozu/ubuntu:20.04

ARG MYSQLSH_VERSION=8.0.24-1

COPY workspace/LICENSE /usr/local/mysql/LICENSE
COPY workspace/bin/mysqlbinlog /usr/local/mysql/bin/mysqlbinlog
COPY workspace/bin/mysql /usrlocal/mysql/bin/mysql

RUN apt-get update \
  && apt-get install -y --no-install-recommends libjemalloc2 zstd python3 libpython3.8 \
  && rm -rf /var/lib/apt/lists/* \
  && curl -o /tmp/mysqlsh.deb -fsL https://dev.mysql.com/get/Downloads/MySQL-Shell/mysql-shell_${MYSQLSH_VERSION}ubuntu20.04_amd64.deb \
  && dpkg -i /tmp/mysqlsh.deb \
  && rm -f /tmp/mysqlsh.deb

ENV PATH=/usr/local/mysql/bin:"$PATH"
USER 10000:10000
