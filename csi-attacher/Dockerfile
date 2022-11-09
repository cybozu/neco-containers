# Stage1: build from source
FROM quay.io/cybozu/golang:1.18-focal AS build

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-attacher
ARG VERSION=4.0.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/kubernetes-csi/external-attacher.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN make

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-attacher
LABEL description="CSI External Attacher"

COPY --from=build ${SRC_DIR}/bin/csi-attacher /
ENTRYPOINT ["/csi-attacher"]
