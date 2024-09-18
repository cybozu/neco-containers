# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-attacher
ARG VERSION=4.6.1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/kubernetes-csi/external-attacher.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN make

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-attacher
LABEL description="CSI External Attacher"

COPY --from=build ${SRC_DIR}/bin/csi-attacher /
ENTRYPOINT ["/csi-attacher"]
