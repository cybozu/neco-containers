# Prometheus container (prometheus, alertmanager, pushgateway)

# Stage1: build from source
FROM quay.io/cybozu/golang:1.20-jammy AS build

ARG PUSHGATEWAY_VERSION=1.6.0

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /work/pushgateway
RUN curl -fsSL -o pushgateway.tar.gz "https://github.com/prometheus/pushgateway/archive/v${PUSHGATEWAY_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f pushgateway.tar.gz \
    && rm -f pushgateway.tar.gz \
    && make build

# Stage2: setup runtime container
FROM scratch

COPY --from=build /work/pushgateway/pushgateway /
COPY --from=build /work/pushgateway/LICENSE /LICENSE

USER 10000:10000

EXPOSE 9091

ENTRYPOINT ["/pushgateway"]
