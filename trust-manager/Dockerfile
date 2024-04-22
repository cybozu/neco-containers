# trust-manager container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

COPY TAG /
WORKDIR /workspace
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL https://github.com/cert-manager/trust-manager/archive/v${VERSION}.tar.gz | \
    tar --strip-components=1 -xzf -

RUN make build

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
