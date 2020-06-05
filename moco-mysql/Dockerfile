FROM ubuntu:18.04

ENV MYSQL_VERSION 8.0.20-1ubuntu18.04
ENV MYSQL_MINOR_VERSION 8.0

ENV MOCO_VERSION 0.1.1

RUN apt-get update && apt-get -y install --no-install-recommends \
      curl \
      openssl \
      ca-certificates \
      libaio1 \
      libmecab2 \
      libnuma1 \
      tzdata \
  && rm -rf /var/lib/apt/lists/*

RUN curl -sSLO https://dev.mysql.com/get/Downloads/MySQL-${MYSQL_MINOR_VERSION}/mysql-common_${MYSQL_VERSION}_amd64.deb \
  && curl -sSLO https://dev.mysql.com/get/Downloads/MySQL-${MYSQL_MINOR_VERSION}/mysql-community-server-core_${MYSQL_VERSION}_amd64.deb \
  && curl -sSLO https://dev.mysql.com/get/Downloads/MySQL-${MYSQL_MINOR_VERSION}/mysql-community-client-core_${MYSQL_VERSION}_amd64.deb \
  && curl -sSLO https://dev.mysql.com/get/Downloads/MySQL-${MYSQL_MINOR_VERSION}/mysql-community-client_${MYSQL_VERSION}_amd64.deb \
  && dpkg -i mysql-common_${MYSQL_VERSION}_amd64.deb \
  && dpkg -i mysql-community-server-core_${MYSQL_VERSION}_amd64.deb \
  && dpkg -i mysql-community-client-core_${MYSQL_VERSION}_amd64.deb \
  && dpkg -i mysql-community-client_${MYSQL_VERSION}_amd64.deb \
  && rm -f mysql-common_${MYSQL_VERSION}_amd64.deb \
  && rm -f mysql-community-server-core_${MYSQL_VERSION}_amd64.deb \
  && rm -f mysql-community-client-core_${MYSQL_VERSION}_amd64.deb \
  && rm -f mysql-community-client_${MYSQL_VERSION}_amd64.deb

RUN groupadd -r mysql \
  && useradd -r -g mysql mysql \
  && mkdir -p /var/lib/mysql \
  && chown -R mysql:mysql /var/lib/mysql

VOLUME /var/lib/mysql

RUN curl -sSL -o /entrypoint https://github.com/cybozu-go/moco/releases/download/v${MOCO_VERSION}/entrypoint \
  && chmod +x /entrypoint

COPY ping.sh /ping.sh

ENTRYPOINT ["mysqld"]
HEALTHCHECK CMD /ping.sh
EXPOSE 3306 33060 33062

USER 10000:10000
