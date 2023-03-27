# victoriametrics-operator container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.20-jammy AS build

ARG VICTORIAMETRICS_OPERATOR_SRCREPO=VictoriaMetrics/operator
ARG VICTORIAMETRICS_OPERATOR_VERSION=0.32.1

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/VictoriaMetrics/operator
RUN curl -fsSL -o victoriametrics-operator.tar.gz "https://github.com/${VICTORIAMETRICS_OPERATOR_SRCREPO}/archive/v${VICTORIAMETRICS_OPERATOR_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f victoriametrics-operator.tar.gz \
    && rm -f victoriametrics-operator.tar.gz \
    && CGO_ENABLED=0 go build -trimpath -ldflags "-w -s -X main.Version=${VICTORIAMETRICS_OPERATOR_VERSION} -X main.BuildData=$(date +%FT%T%z)" -o bin/manager main.go

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/src/github.com/VictoriaMetrics/operator/bin/manager /
COPY --from=build /go/src/github.com/VictoriaMetrics/operator/LICENSE /

USER 10000:10000

ENTRYPOINT ["/manager"]
