# Stage1: build from source
FROM quay.io/cybozu/golang:1.20-focal AS build

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-resizer
ARG VERSION=1.9.2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/kubernetes-csi/external-resizer.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN make

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-resizer
LABEL description="CSI External Resizer"

COPY --from=build ${SRC_DIR}/bin/csi-resizer /
ENTRYPOINT ["/csi-resizer"]
