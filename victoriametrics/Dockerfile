# victoriametrics-xxx container

ARG VICTORIAMETRICS_SRCREPO=VictoriaMetrics/VictoriaMetrics
ARG VICTORIAMETRICS_SINGLE_VERSION=1.91.3
ARG VICTORIAMETRICS_CLUSTER_VERSION=1.91.3-cluster

# Stage1: build from source
FROM quay.io/cybozu/golang:1.20-jammy AS build-single

ARG VICTORIAMETRICS_SRCREPO
ARG VICTORIAMETRICS_SINGLE_VERSION

COPY download_and_build.sh /
WORKDIR /go/src/github.com/VictoriaMetrics/VictoriaMetrics
RUN VICTORIAMETRICS_VERSION=${VICTORIAMETRICS_SINGLE_VERSION} /download_and_build.sh victoria-metrics vmagent vmalert vmbackup vmrestore vmctl

#
FROM quay.io/cybozu/golang:1.20-jammy AS build-cluster

ARG VICTORIAMETRICS_SRCREPO
ARG VICTORIAMETRICS_CLUSTER_VERSION

COPY download_and_build.sh /
WORKDIR /go/src/github.com/VictoriaMetrics/VictoriaMetrics
RUN VICTORIAMETRICS_VERSION=${VICTORIAMETRICS_CLUSTER_VERSION} /download_and_build.sh vminsert vmselect vmstorage


# Stage2: setup runtime containers
FROM quay.io/cybozu/ubuntu:22.04 AS vmsingle

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/victoria-metrics /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8428

ENTRYPOINT ["/victoria-metrics"]

#
FROM quay.io/cybozu/ubuntu:22.04 AS vmagent

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmagent /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8429

ENTRYPOINT ["/vmagent"]

#
FROM quay.io/cybozu/ubuntu:22.04 AS vmalert

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmalert /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8880

ENTRYPOINT ["/vmalert"]

#
FROM quay.io/cybozu/ubuntu:22.04 AS vmbackup

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmbackup /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000

ENTRYPOINT ["/vmbackup"]

#
FROM quay.io/cybozu/ubuntu:22.04 AS vmrestore

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmrestore /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000

ENTRYPOINT ["/vmrestore"]

#
FROM quay.io/cybozu/ubuntu:22.04 AS vmctl

COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmctl /
COPY --from=build-single /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000

ENTRYPOINT ["/vmctl"]

#
FROM quay.io/cybozu/ubuntu:22.04 AS vminsert

COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vminsert /
COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8480

ENTRYPOINT ["/vminsert"]

#
FROM quay.io/cybozu/ubuntu:22.04 AS vmselect

COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmselect /
COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8481

ENTRYPOINT ["/vmselect"]

#
FROM quay.io/cybozu/ubuntu:22.04 AS vmstorage

COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/bin/vmstorage /
COPY --from=build-cluster /go/src/github.com/VictoriaMetrics/VictoriaMetrics/LICENSE /

USER 10000:10000
EXPOSE 8482
EXPOSE 8400
EXPOSE 8401

ENTRYPOINT ["/vmstorage"]
