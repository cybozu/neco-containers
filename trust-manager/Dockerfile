# trust-manager container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.23-jammy AS build

COPY TAG /
WORKDIR /workspace
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL https://github.com/cert-manager/trust-manager/archive/v${VERSION}.tar.gz | \
    tar --strip-components=1 -xzf -

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
