# metallb container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.16-focal AS build

## we use the master branch because there has been
## no release since Jul. 2019.
ARG MYSQLD_EXPORTER_VERSION=master

RUN git clone -b ${MYSQLD_EXPORTER_VERSION} --depth 1 https://github.com/prometheus/mysqld_exporter \
    && make -C mysqld_exporter build

# Stage2: setup runtime container
FROM scratch

COPY --from=build /work/mysqld_exporter/LICENSE /LICENSE
COPY --from=build /work/mysqld_exporter/mysqld_exporter /mysqld_exporter

USER 10000:10000
EXPOSE 9104

ENTRYPOINT ["/mysqld_exporter"]
