# Rook container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.19-focal AS build

ARG ROOK_VERSION=1.11.4
ARG ROOK_DIR=/work/go/src/github.com/rook/rook

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV GOPATH=/work/go
ENV PATH $PATH:${ROOK_DIR}/.cache/tools/linux_amd64
RUN git clone https://github.com/rook/rook.git ${ROOK_DIR}

WORKDIR ${ROOK_DIR}

# Install kubectl binary as required for generating csv
# ref: https://github.com/rook/rook/commit/972202db0d3a57863d2a1ebac2047f99d169d0a9
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x ./kubectl \
    && mv ./kubectl /usr/local/bin/kubectl

RUN mkdir -p ${ROOK_DIR}/.cache/tools/linux_amd64
# Make a symbolic link of helm to allow the build script to use `helm` command without specifying the version.
RUN HELM_VERSION=$(grep "^HELM_VERSION"  ${ROOK_DIR}/build/makelib/helm.mk | cut -d " " -f 3) && \
    ln -s ${ROOK_DIR}/.cache/tools/linux_amd64/helm-${HELM_VERSION} ${ROOK_DIR}/.cache/tools/linux_amd64/helm

RUN git checkout v${ROOK_VERSION}
RUN mkdir -p /tmp/rook
# Don't delete temporary directory to save output artifacts
RUN sed -i -e '/@rm -fr $(TEMP)/d' images/ceph/Makefile
RUN make build IMAGES="ceph" BUILD_CONTAINER_IMAGE=false TEMP=/tmp/rook
# Copy output artifacts to root directory to minimize differences between Stage2 and upstream Dockerfile.
RUN cp -r /tmp/rook/* /

# Stage2: setup runtime container
FROM quay.io/cybozu/ceph:17.2.6.1

ARG ROOK_DIR=/work/go/src/github.com/rook/rook
COPY --from=build ${ROOK_DIR}/LICENSE /usr/local/rook/LICENSE

# Followings are based on upstream Dockerfile
# Note: s5cmd is not installed because teleport node image has s3cmd.
COPY --from=build rook toolbox.sh set-ceph-debug-level /usr/local/bin/
COPY --from=build ceph-monitoring /etc/ceph-monitoring
COPY --from=build rook-external /etc/rook-external/
COPY --from=build ceph-csv-templates /etc/ceph-csv-templates

# create or modify owner and permissions to make a watch-active container of a MGR work properly
RUN groupadd rook -g 2016 && \
    useradd rook -u 2016 -g rook
USER rook:rook

ENTRYPOINT ["/usr/local/bin/rook"]
