# victoriametrics-xxx container

ARG VICTORIAMETRICS_SRCREPO=VictoriaMetrics/VictoriaMetrics
ARG VICTORIAMETRICS_SINGLE_VERSION=1.111.0
ARG VICTORIAMETRICS_CLUSTER_VERSION=1.111.0-cluster

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.23-jammy AS build-single

ARG VICTORIAMETRICS_SRCREPO
ARG VICTORIAMETRICS_SINGLE_VERSION

COPY download_and_build.sh /
WORKDIR /go/src/github.com/VictoriaMetrics/VictoriaMetrics
RUN VICTORIAMETRICS_VERSION=${VICTORIAMETRICS_SINGLE_VERSION} /download_and_build.sh victoria-metrics vmagent vmalert vmbackup vmrestore vmctl

#
FROM ghcr.io/cybozu/golang:1.23-jammy AS build-cluster

ARG VICTORIAMETRICS_SRCREPO
ARG VICTORIAMETRICS_CLUSTER_VERSION

COPY download_and_build.sh /
WORKDIR /go/src/github.com/VictoriaMetrics/VictoriaMetrics
RUN VICTORIAMETRICS_VERSION=${VICTORIAMETRICS_CLUSTER_VERSION} /download_and_build.sh vminsert vmselect vmstorage


# Stage2: setup runtime containers
FROM ghcr.io/cybozu/ubuntu:22.04 AS vmsingle
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/victoria-metrics /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8428

ENTRYPOINT ["/victoria-metrics"]

#
FROM ghcr.io/cybozu/ubuntu:22.04 AS vmagent
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmagent /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8429

ENTRYPOINT ["/vmagent"]

#
FROM ghcr.io/cybozu/ubuntu:22.04 AS vmalert
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmalert /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8880

ENTRYPOINT ["/vmalert"]

#
FROM ghcr.io/cybozu/ubuntu:22.04 AS vmbackup
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmbackup /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000

ENTRYPOINT ["/vmbackup"]

#
FROM ghcr.io/cybozu/ubuntu:22.04 AS vmrestore
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmrestore /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000

ENTRYPOINT ["/vmrestore"]

#
FROM ghcr.io/cybozu/ubuntu:22.04 AS vmctl
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmctl /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000

ENTRYPOINT ["/vmctl"]

#
FROM ghcr.io/cybozu/ubuntu:22.04 AS vminsert
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vminsert /
COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8480

ENTRYPOINT ["/vminsert"]

#
FROM ghcr.io/cybozu/ubuntu:22.04 AS vmselect
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmselect /
COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8481

ENTRYPOINT ["/vmselect"]

#
FROM ghcr.io/cybozu/ubuntu:22.04 AS vmstorage
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmstorage /
COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8482
EXPOSE 8400
EXPOSE 8401

ENTRYPOINT ["/vmstorage"]
