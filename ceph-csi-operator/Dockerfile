ARG SRC_DIR=/work/go/src/github.com/ceph/ceph-csi-operator
# Build the manager binary
FROM ghcr.io/cybozu/golang:1.24-noble AS builder

ARG SRC_DIR
ARG VERSION=0.4.1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone -b v${VERSION} --depth=1 https://github.com/ceph/ceph-csi-operator.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

# Build
RUN CGO_ENABLED=0 go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
ARG SRC_DIR
WORKDIR /
COPY --from=builder ${SRC_DIR}/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
