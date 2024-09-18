# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/node-driver-registrar
ARG VERSION=2.11.1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/kubernetes-csi/node-driver-registrar.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN make

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/node-driver-registrar
LABEL description="CSI Node driver registrar"

COPY --from=build ${SRC_DIR}/bin/csi-node-driver-registrar /
ENTRYPOINT ["/csi-node-driver-registrar"]
