# Stage1: build from source
FROM quay.io/cybozu/golang:1.19-focal AS build

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/node-driver-registrar
ARG VERSION=2.7.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/kubernetes-csi/node-driver-registrar.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN make

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/node-driver-registrar
LABEL description="CSI Node driver registrar"

COPY --from=build ${SRC_DIR}/bin/csi-node-driver-registrar /
ENTRYPOINT ["/csi-node-driver-registrar"]
