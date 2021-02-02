FROM quay.io/cybozu/golang:1.15-focal AS build
ARG LOKI_VERSION=2.1.0

WORKDIR /work/loki
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN git clone --depth=1 -b v${LOKI_VERSION} https://github.com/grafana/loki /work/loki

RUN make BUILD_IN_CONTAINER=false loki logcli

FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /work/loki/cmd/loki/loki /usr/bin/loki
COPY --from=build /work/loki/cmd/logcli/logcli /usr/bin/logcli
COPY --from=build /work/loki/cmd/loki/loki-docker-config.yaml /etc/loki/local-config.yaml

RUN mkdir -p /loki/rules && \
    mkdir -p /loki/tmprules && \
    chown -R 10000:10000 /etc/loki /loki

USER 10000:10000
EXPOSE 3100
ENTRYPOINT [ "/usr/bin/loki" ]
CMD ["-config.file=/etc/loki/local-config.yaml"]
