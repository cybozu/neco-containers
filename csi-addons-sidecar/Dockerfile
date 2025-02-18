ARG SRC_DIR=/work/go/src/github.com/csi-addons/kubernetes-csi-addons
# Build the sidecar binary
FROM ghcr.io/cybozu/golang:1.22-jammy AS builder

ARG SRC_DIR
ARG VERSION=0.11.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/csi-addons/kubernetes-csi-addons.git ${SRC_DIR}

ENV CGO_ENABLED=0
WORKDIR ${SRC_DIR}

# Build the sidecar and csi-addons tool for admin usage and testing
RUN make build

# Use distroless as minimal base image to package the sidecar binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# DL3007 ignored as it asks to pin the version explicitly to a release tag
# hadolint ignore=DL3007
FROM gcr.io/distroless/static:latest
ARG SRC_DIR
WORKDIR /
COPY --from=builder ${SRC_DIR}/bin/csi-addons-sidecar /usr/sbin/
COPY --from=builder ${SRC_DIR}/bin/csi-addons /usr/bin/

ENTRYPOINT ["/usr/sbin/csi-addons-sidecar"]
