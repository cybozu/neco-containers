# trust-manager container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build
ENV TRUST_MANAGER_HASH=c1ba3525637adc858a8addb24b87be4d0189a77c0e532734d9aeac41b3351f71

COPY TAG /
WORKDIR /workspace
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d \. -f1-3 < /TAG ) \
    && curl -fsSL https://github.com/cert-manager/trust-manager/archive/v${VERSION}.tar.gz -o /tmp/trust-manager.tar.gz \
    && echo "${TRUST_MANAGER_HASH} /tmp/trust-manager.tar.gz" | sha256sum -c - \
    && tar zxf /tmp/trust-manager.tar.gz --strip-components=1 -C /workspace/

# Parameters retrieved from:
# $ make ko-config-manager
# $ cat _bin/scratch/image/oci-layout-manager.v0.13.0.ko_config.yaml
RUN CGO_ENABLED=0 GO_EXPERIMENT= go build -ldflags '-w -s' -o bin/trust-manager ./cmd/trust-manager 

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"
LABEL description="trust-manager is an operator for distributing trust bundles across a Kubernetes cluster"

WORKDIR /
COPY --from=build /workspace/bin/trust-manager /usr/bin/trust-manager
COPY --from=build /workspace/LICENSE /LICENSE

USER 10000:10000
EXPOSE 6060 6443 9402
ENTRYPOINT ["/usr/bin/trust-manager"]
