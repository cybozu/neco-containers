# cert-manager container

FROM quay.io/cybozu/ubuntu:20.04

COPY workspace/webhook /usr/local/bin/webhook
COPY workspace/cainjector /usr/local/bin/cainjector
COPY workspace/controller /usr/local/bin/controller
COPY workspace/LICENSE /usr/local/share/doc/cert-manager/LICENSE

EXPOSE 9402

USER 10000:10000
