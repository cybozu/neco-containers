# ==================== pyroscope-builder image ====================
# https://github.com/grafana/pyroscope/blob/main/cmd/pyroscope/Dockerfile
FROM ghcr.io/cybozu/golang:1.24-noble AS builder

COPY src/pyroscope ./pyroscope
RUN make -C pyroscope EMBEDASSETS="" go/bin

# ==================== pyroscope image ====================
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/pyroscope/LICENSE /usr/local/pyroscope/LICENSE
COPY --from=builder /work/pyroscope/cmd/pyroscope/pyroscope.yaml /etc/pyroscope/config.yaml
COPY --from=builder /work/pyroscope/profilecli /usr/local/bin/profilecli
COPY --from=builder /work/pyroscope/pyroscope /usr/local/bin/pyroscope

RUN mkdir -p /data && \
    mkdir -p /data-compactor && \
    chown -R 10000:10000 /data && \
    chown -R 10000:10000 /data-compactor

USER 10000:10000
EXPOSE 4040
ENTRYPOINT [ "/usr/local/bin/pyroscope" ]
CMD ["-config.file=/etc/pyroscope/config.yaml"]
