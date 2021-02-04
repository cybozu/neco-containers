FROM quay.io/cybozu/golang:1.15-focal AS build
ARG LOKI_VERSION=2.1.0

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone --depth=1 -b v${LOKI_VERSION} https://github.com/grafana/loki /work/loki
RUN make -C /work/loki BUILD_IN_CONTAINER=false touch-protos loki logcli

FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /work/loki/cmd/loki/loki /usr/local/loki/bin/loki
COPY --from=build /work/loki/cmd/logcli/logcli /usr/local/loki/bin/logcli
COPY --from=build /work/loki/LICENSE /usr/local/loki/LICENSE
COPY --from=build /work/loki/cmd/loki/loki-docker-config.yaml /usr/local/loki/etc/local-config.yaml

RUN mkdir -p /loki/rules && \
    mkdir -p /loki/tmprules && \
    chown -R 10000:10000 /loki

USER 10000:10000
EXPOSE 3100
ENV PATH=/usr/local/loki/bin:"$PATH"
VOLUME /loki
ENTRYPOINT [ "/usr/local/loki/bin/loki" ]
CMD ["-config.file=/usr/local/loki/etc/local-config.yaml"]
