# CSI sidecar container image

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG EXTERNAL_PROVISIONER_VERSION=1.1.1
ARG NODE_DRIVER_REGISTRAR_VERSION=1.1.0
ARG EXTERNAL_ATTACHER_VERSION=1.1.1

RUN mkdir -p /go/src/github.com/kubernetes-csi && \
    curl -sSLf https://github.com/kubernetes-csi/external-provisioner/archive/v${EXTERNAL_PROVISIONER_VERSION}.tar.gz | \
        tar zxf - -C /go/src/github.com/kubernetes-csi/ && \
    mv /go/src/github.com/kubernetes-csi/external-provisioner-${EXTERNAL_PROVISIONER_VERSION} /go/src/github.com/kubernetes-csi/external-provisioner/

RUN curl -sSLf https://github.com/kubernetes-csi/node-driver-registrar/archive/v${NODE_DRIVER_REGISTRAR_VERSION}.tar.gz | \
        tar zxf - -C /go/src/github.com/kubernetes-csi/ && \
    mv /go/src/github.com/kubernetes-csi/node-driver-registrar-${NODE_DRIVER_REGISTRAR_VERSION} /go/src/github.com/kubernetes-csi/node-driver-registrar/

RUN curl -sSLf https://github.com/kubernetes-csi/external-attacher/archive/v${EXTERNAL_ATTACHER_VERSION}.tar.gz | \
        tar zxf - -C /go/src/github.com/kubernetes-csi/ && \
    mv /go/src/github.com/kubernetes-csi/external-attacher-${EXTERNAL_ATTACHER_VERSION} /go/src/github.com/kubernetes-csi/external-attacher/

WORKDIR /go/src/github.com/kubernetes-csi
RUN for pkg in external-provisioner node-driver-registrar external-attacher; do (cd ${pkg}; make); done

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/src/github.com/kubernetes-csi/external-provisioner/bin/csi-provisioner /usr/local/csi-sidecar/bin/csi-provisioner
COPY --from=build /go/src/github.com/kubernetes-csi/external-provisioner/LICENSE  /usr/local/csi-sidecar/LICENSE

COPY --from=build /go/src/github.com/kubernetes-csi/node-driver-registrar/bin/csi-node-driver-registrar /usr/local/csi-sidecar/bin/csi-node-driver-registrar

COPY --from=build /go/src/github.com/kubernetes-csi/external-attacher/bin/csi-attacher /usr/local/csi-sidecar/bin/csi-attacher

ENV PATH=/usr/local/csi-sidecar/bin:"$PATH"

USER 10000:10000
