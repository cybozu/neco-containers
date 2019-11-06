# cert-manager container

FROM quay.io/cybozu/ubuntu:18.04

COPY webhook /usr/local/bin/webhook
COPY cainjector /usr/local/bin/cainjector
COPY controller /usr/local/bin/controller
COPY LICENSE /usr/local/share/doc/cert-manager/LICENSE

EXPOSE 9402

USER 10000:10000
