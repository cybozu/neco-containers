FROM quay.io/cybozu/ubuntu:18.04

COPY bin/local-pv-provisioner /local-pv-provisioner

USER 10000:10000

ENTRYPOINT ["/local-pv-provisioner"]
