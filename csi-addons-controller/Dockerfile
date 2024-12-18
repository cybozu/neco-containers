ARG SRC_DIR=/work/go/src/github.com/csi-addons/kubernetes-csi-addons
# Build the manager binary
FROM ghcr.io/cybozu/golang:1.22-jammy AS builder

ARG SRC_DIR
ARG VERSION=0.11.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/csi-addons/kubernetes-csi-addons.git ${SRC_DIR}

ENV CGO_ENABLED=0
WORKDIR ${SRC_DIR}

# Build
RUN make build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
ARG SRC_DIR
WORKDIR /
COPY --from=builder ${SRC_DIR}/bin/csi-addons-manager .
USER 65532:65532

ENTRYPOINT ["/csi-addons-manager"]
