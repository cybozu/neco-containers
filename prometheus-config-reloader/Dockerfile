# prometheus-config-reloader container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.15-bionic AS build

ARG PROMETHEUS_OPERATOR_VERSION=0.43.2

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/prometheus-operator/prometheus-operator
RUN curl -fsSL -o prometheus-operator.tar.gz "https://github.com/prometheus-operator/prometheus-operator/archive/v${PROMETHEUS_OPERATOR_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f prometheus-operator.tar.gz \
    && rm -f prometheus-operator.tar.gz \
    && go install ./cmd/prometheus-config-reloader

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin/prometheus-config-reloader /bin
COPY --from=build /go/src/github.com/prometheus-operator/prometheus-operator/LICENSE /
COPY --from=build /go/src/github.com/prometheus-operator/prometheus-operator/NOTICE /

USER 10000:10000

ENTRYPOINT ["/bin/prometheus-config-reloader"]
