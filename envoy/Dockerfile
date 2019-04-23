# Envoy container

# Stage1: load official container image 
FROM envoyproxy/envoy:v1.9.1 AS official

# Stage2: set up runtime container
FROM quay.io/cybozu/ubuntu:18.04

RUN mkdir -p /etc/envoy

COPY --from=official /usr/local/bin/envoy /usr/local/bin/envoy
COPY --from=official /etc/envoy/envoy.yaml /etc/envoy/envoy.yaml

# TODO: Uncomment the line below after envoy is updated to 1.10.0+
#COPY --from=official /docker-entrypoint.sh /docker-entrypoint.sh
RUN curl -sSLf -o /docker-entrypoint.sh https://github.com/envoyproxy/envoy/raw/v1.10.0/ci/docker-entrypoint.sh \
    && chmod +x /docker-entrypoint.sh \
    && curl -sSLf -o /usr/local/bin/LICENSE https://github.com/envoyproxy/envoy/raw/master/LICENSE 

EXPOSE 10000

USER 10000:10000
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]
