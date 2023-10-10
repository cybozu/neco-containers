FROM quay.io/cybozu/golang:1.20-focal as builder

WORKDIR /work

COPY . .

RUN go build -o ceph-extra-exporter

FROM quay.io/cybozu/ceph:17.2.6.1

COPY --from=builder /work/ceph-extra-exporter /

USER 1001:1001
EXPOSE 8080/tcp

ENTRYPOINT [ "/ceph-extra-exporter" ]
