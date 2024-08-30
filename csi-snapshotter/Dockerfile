# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-snapshotter
ARG VERSION=7.0.2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/kubernetes-csi/external-snapshotter.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN make

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-snapshotter
LABEL description="CSI External Snapshotter Sidecar"

COPY --from=build ${SRC_DIR}/bin/csi-snapshotter /
ENTRYPOINT ["/csi-snapshotter"]
