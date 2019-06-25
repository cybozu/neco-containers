# Envoy container

FROM quay.io/cybozu/ubuntu:18.04

COPY workspace/envoy /usr/local/bin/envoy
COPY workspace/docker-entrypoint.sh /docker-entrypoint.sh
COPY workspace/LICENSE /usr/local/share/doc/envoy/LICENSE
COPY envoy.yaml /etc/envoy/envoy.yaml

EXPOSE 9901

USER 10000:10000
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]
