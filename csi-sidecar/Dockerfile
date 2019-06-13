# CSI sidecar container image

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build

# Install dep
ENV DEP_VERSION=0.5.0
RUN curl -sSLf https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 -o /usr/local/bin/dep && \
    chmod +x /usr/local/bin/dep

ARG EXTERNAL_PROVISIONER_VERSION=1.1.1
ARG NODE_DRIVER_REGISTRAR_VERSION=1.1.0
ARG EXTERNAL_ATTACHER_VERSION=1.1.1


RUN curl -sSLf https://github.com/kubernetes-csi/external-provisioner/archive/v${EXTERNAL_PROVISIONER_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mv /work/external-provisioner-${EXTERNAL_PROVISIONER_VERSION} /work/external-provisioner/

RUN mkdir -p ${GOPATH}/src/dummy \
    && cp /work/external-provisioner/Gopkg.toml ${GOPATH}/src/dummy/Gopkg.toml \
    && cp /work/external-provisioner/Gopkg.lock ${GOPATH}/src/dummy/Gopkg.lock

RUN cd ${GOPATH}/src/dummy && \
    dep ensure -vendor-only && \
    mv vendor/* ${GOPATH}/src/ && \
    rmdir vendor

WORKDIR /work/external-provisioner/
RUN make && rm -rf ${GOPATH}/src/dummy


RUN curl -sSLf https://github.com/kubernetes-csi/node-driver-registrar/archive/v${NODE_DRIVER_REGISTRAR_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mv /work/node-driver-registrar-${NODE_DRIVER_REGISTRAR_VERSION} /work/node-driver-registrar/

RUN mkdir -p ${GOPATH}/src/dummy \
    && cp /work/node-driver-registrar/Gopkg.toml ${GOPATH}/src/dummy/Gopkg.toml \
    && cp /work/node-driver-registrar/Gopkg.lock ${GOPATH}/src/dummy/Gopkg.lock

RUN cd ${GOPATH}/src/dummy && \
    dep ensure -vendor-only && \
    mv vendor/* ${GOPATH}/src/ && \
    rmdir vendor

WORKDIR /work/node-driver-registrar/
RUN make && rm -rf ${GOPATH}/src/dummy

RUN curl -sSLf https://github.com/kubernetes-csi/external-attacher/archive/v${EXTERNAL_ATTACHER_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mv /work/external-attacher-${EXTERNAL_ATTACHER_VERSION} /work/external-attacher/

RUN mkdir -p ${GOPATH}/src/dummy \
    && cp /work/external-attacher/Gopkg.toml ${GOPATH}/src/dummy/Gopkg.toml \
    && cp /work/external-attacher/Gopkg.lock ${GOPATH}/src/dummy/Gopkg.lock

RUN cd ${GOPATH}/src/dummy && \
    dep ensure -vendor-only && \
    mv vendor/* ${GOPATH}/src/ && \
    rmdir vendor

WORKDIR /work/external-attacher/
RUN make && rm -rf ${GOPATH}/src/dummy

# Stage2: setup runtime container
FROM scratch

COPY --from=build /work/external-provisioner/bin/csi-provisioner /usr/local/csi-sidecar/bin/csi-provisioner
COPY --from=build /work/external-provisioner/LICENSE  /usr/local/csi-sidecar/bin/LICENSE

COPY --from=build /work/node-driver-registrar/bin/csi-node-driver-registrar /usr/local/csi-sidecar/bin/csi-node-driver-registrar

COPY --from=build /work/external-attacher/bin/csi-attacher /usr/local/csi-sidecar/bin/csi-attacher

ENV PATH=/usr/local/csi-sidecar/bin:"$PATH"

USER 10000:10000
