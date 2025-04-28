FROM ghcr.io/cybozu/golang:1.23-noble AS builder

WORKDIR /work

COPY . .

RUN go build -o ceph-extra-exporter

FROM ghcr.io/cybozu/ceph:19.2.2.1
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/ceph-extra-exporter /

USER 1001:1001
EXPOSE 8080/tcp

ENTRYPOINT [ "/ceph-extra-exporter" ]
