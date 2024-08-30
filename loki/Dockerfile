FROM ghcr.io/cybozu/golang:1.22-jammy AS build
ARG LOKI_VERSION=2.9.9

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone --depth=1 -b v${LOKI_VERSION} https://github.com/grafana/loki /work/loki
COPY log-per-chunk-flush.patch /work
WORKDIR /work/loki
RUN patch -p1 < ../log-per-chunk-flush.patch
RUN make BUILD_IN_CONTAINER=false loki logcli loki-canary

FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/loki/cmd/loki/loki /usr/local/loki/bin/loki
COPY --from=build /work/loki/cmd/logcli/logcli /usr/local/loki/bin/logcli
COPY --from=build /work/loki/cmd/loki-canary/loki-canary /usr/local/loki/bin/loki-canary
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
