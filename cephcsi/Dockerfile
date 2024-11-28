# ceph-csi container
ARG SRC_DIR="/work/go/src/github.com/ceph/ceph-csi/"
ARG GO_ARCH=amd64
ARG BASE_IMAGE="ghcr.io/cybozu/ceph:18.2.4.2"

FROM ${BASE_IMAGE} as updated_base

RUN mkdir /etc/selinux || true && touch /etc/selinux/config

FROM updated_base as builder

LABEL stage="build"

ARG CSI_IMAGE_NAME=ghcr.io/cybozu/cephcsi
ARG CSI_IMAGE_VERSION=3.12.3
ARG GO_ARCH
ARG SRC_DIR
ARG GIT_COMMIT
ARG GOROOT=/usr/local/go

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN apt-get update && apt-get -y install git

RUN git clone -b v${CSI_IMAGE_VERSION} --depth=1 https://github.com/ceph/ceph-csi.git ${SRC_DIR}

WORKDIR ${SRC_DIR}

RUN cp build.env /

RUN source /build.env && \
    ( test -n "${GO_ARCH}" && exit 0; echo -e "\n\nMissing GO_ARCH argument for building image, install Golang or run: make image-cephcsi GOARCH=amd64\n\n"; exit 1 ) && \
    mkdir -p ${GOROOT} && \
    curl https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-${GO_ARCH}.tar.gz | tar xzf - -C ${GOROOT} --strip-components=1

# test if the downloaded version of Golang works (different arch?)
RUN ${GOROOT}/bin/go version && ${GOROOT}/bin/go env

RUN apt-get update && apt-get -y install \
    gcc \
    make \
    && true

ENV GOROOT=${GOROOT} \
    GOPATH=/work/go \
    CGO_ENABLED=1 \
    GIT_COMMIT="${GIT_COMMIT}" \
    ENV_CSI_IMAGE_VERSION="${CSI_IMAGE_VERSION}" \
    ENV_CSI_IMAGE_NAME="${CSI_IMAGE_NAME}" \
    PATH="${GOROOT}/bin:${GOPATH}/bin:${PATH}"

# Build executable
RUN make cephcsi

#-- Final container
FROM updated_base
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG SRC_DIR

LABEL version=${CSI_IMAGE_VERSION} \
    architecture=${GO_ARCH} \
    description="Ceph-CSI Plugin"

COPY --from=builder ${SRC_DIR}/_output/cephcsi /usr/local/bin/cephcsi

# verify that all dynamically linked libraries are available
RUN [ $(ldd /usr/local/bin/cephcsi | grep -c '=> not found') = '0' ]

ENTRYPOINT ["/usr/local/bin/cephcsi"]
