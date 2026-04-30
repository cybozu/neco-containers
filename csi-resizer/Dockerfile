# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.25-noble AS build

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-resizer
ARG VERSION=2.1.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/kubernetes-csi/external-resizer.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN make

# Stage2: setup runtime container
FROM gcr.io/distroless/static:latest
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG SRC_DIR=/work/go/src/github.com/kubernetes-csi/external-resizer
LABEL description="CSI External Resizer"

COPY --from=build ${SRC_DIR}/bin/csi-resizer /
ENTRYPOINT ["/csi-resizer"]
