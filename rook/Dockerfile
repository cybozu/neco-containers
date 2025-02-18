# Rook container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG ROOK_VERSION=1.16.3
ARG ROOK_DIR=/work/go/src/github.com/rook/rook

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV GOPATH=/work/go
ENV PATH $PATH:${ROOK_DIR}/.cache/tools/linux_amd64
RUN git clone https://github.com/rook/rook.git ${ROOK_DIR}

WORKDIR ${ROOK_DIR}

RUN mkdir -p ${ROOK_DIR}/.cache/tools/linux_amd64
# Make a symbolic link of helm to allow the build script to use `helm` command without specifying the version.
RUN HELM_VERSION=$(grep "^HELM_VERSION"  ${ROOK_DIR}/build/makelib/helm.mk | cut -d " " -f 3) && \
    ln -s ${ROOK_DIR}/.cache/tools/linux_amd64/helm-${HELM_VERSION} ${ROOK_DIR}/.cache/tools/linux_amd64/helm

RUN git checkout v${ROOK_VERSION}
RUN mkdir -p /tmp/rook
RUN make build IMAGES="ceph" BUILD_CONTAINER_IMAGE=false BUILD_CONTEXT_DIR=/tmp/rook SAVE_BUILD_CONTEXT_DIR=true
# Copy output artifacts to root directory to minimize differences between Stage2 and upstream Dockerfile.
RUN cp -r /tmp/rook/* /

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ceph:18.2.4.2
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

ARG ROOK_DIR=/work/go/src/github.com/rook/rook
COPY --from=build ${ROOK_DIR}/LICENSE /usr/local/rook/LICENSE

# Followings are based on upstream Dockerfile
# Note: s5cmd is not installed because teleport node image has s3cmd.
RUN apt-get update && apt-get install -y --no-install-recommends iproute2 && rm -rf /var/lib/apt/lists/*
COPY --from=build rook toolbox.sh set-ceph-debug-level /usr/local/bin/
COPY --from=build ceph-monitoring /etc/ceph-monitoring
COPY --from=build rook-external /etc/rook-external/

# create or modify owner and permissions to make a watch-active container of a MGR work properly
RUN groupadd rook -g 2016 && \
    useradd rook -u 2016 -g rook
USER rook:rook

ENTRYPOINT ["/usr/local/bin/rook"]
