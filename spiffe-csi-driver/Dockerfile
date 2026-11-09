# spiffe-csi container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.3.2_noble@sha256:d78937fddb26c4350824d4022f5c9dfe20bf762811ea31065f8c6d467c58bc30 AS build

COPY TAG /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d . -f 1,2,3 < /TAG) \
    && git clone --depth=1 -b v${VERSION} https://github.com/spiffe/spiffe-csi.git

WORKDIR /work/spiffe-csi

RUN make build && \
    mv bin/spiffe-csi-driver /work/spiffe-csi-driver

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/spiffe-csi-driver /usr/local/bin/spiffe-csi-driver
COPY --from=build /work/spiffe-csi/LICENSE /usr/local/share/doc/spiffe-csi/LICENSE

ENTRYPOINT ["/usr/local/bin/spiffe-csi-driver"]
