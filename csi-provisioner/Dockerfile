# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.20-focal AS build

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-provisioner
ARG VERSION=3.6.2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/kubernetes-csi/external-provisioner.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN make

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:20.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-provisioner
LABEL description="CSI External Provisioner"

COPY --from=build ${SRC_DIR}/bin/csi-provisioner /
ENTRYPOINT ["/csi-provisioner"]
