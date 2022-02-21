# Rook container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.16-focal AS build

ARG ROOK_VERSION=1.8.3
# These version parameters are copied from rook/build/makelib/golang.mk and helm.mk
ARG HELM_VERSION=3.6.2
ARG YQ_VERSION=4.14.2
ARG ROOK_DIR=/work/go/src/github.com/rook/rook
ARG USE_CEPH_WITH_ASAN="OFF"

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV GOPATH=/work/go
ENV PATH $PATH:${ROOK_DIR}/.cache/tools/linux_amd64
RUN git clone https://github.com/rook/rook.git ${ROOK_DIR}

WORKDIR ${ROOK_DIR}

RUN mkdir -p ${ROOK_DIR}/.cache/tools/linux_amd64
RUN ln -s ${ROOK_DIR}/.cache/tools/linux_amd64/yq-v${YQ_VERSION} ${ROOK_DIR}/.cache/tools/linux_amd64/yq
RUN ln -s ${ROOK_DIR}/.cache/tools/linux_amd64/helm-v${HELM_VERSION} ${ROOK_DIR}/.cache/tools/linux_amd64/helm

RUN git checkout v${ROOK_VERSION}
COPY rook_asan.patch /tmp/
RUN if [ "$USE_CEPH_WITH_ASAN" = "ON" ]; then \
        echo "USE_CEPH_WITH_ASAN is ON. rook_asan.patch will be applied."; \
        git apply /tmp/rook_asan.patch; \
    fi
RUN make build IMAGES="ceph" BUILD_CONTAINER_IMAGE=false

WORKDIR ${ROOK_DIR}/images/ceph
RUN mkdir /tmp/csv_template_dir
RUN make CSV_TEMPLATE_DIR=/tmp/csv_template_dir generate-csv-templates

# Stage2: setup runtime container
FROM quay.io/cybozu/ceph:16.2.7.0

ARG ROOK_DIR=/work/go/src/github.com/rook/rook

COPY --from=build ${ROOK_DIR}/_output/bin/linux_amd64/rook \
    ${ROOK_DIR}/images/ceph/toolbox.sh \
    ${ROOK_DIR}/images/ceph/set-ceph-debug-level \
    /usr/local/bin/

COPY --from=build ${ROOK_DIR}/deploy/examples/monitoring /etc/ceph-monitoring
COPY --from=build ${ROOK_DIR}/deploy/examples/create-external-cluster-resources.* /etc/rook-external/
COPY --from=build ${ROOK_DIR}/tests/ceph-status-out /etc/rook-external/test-data/
COPY --from=build /tmp/csv_template_dir/deploy/olm/templates /etc/ceph-csv-templates
COPY --from=build ${ROOK_DIR}/.cache/crds/* /etc/ceph-csv-templates/crds/
COPY --from=build ${ROOK_DIR}/LICENSE /usr/local/rook/LICENSE

# The Rook Operator and the toolbox run under a rook user rather than root from v1.8.0
RUN groupadd rook -g 2016 && \
    useradd rook -u 2016 -g rook

# create or modify owner and permissions to make a watch-active container of a MGR work properly
RUN install -d -o rook -g rook -m 0700 /var/lib/rook
RUN chown rook:rook /etc/ceph

USER rook:rook

ENTRYPOINT ["/usr/local/bin/rook"]
