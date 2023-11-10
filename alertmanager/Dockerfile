# Alertmanager container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.21-jammy AS build

ARG ALERTMANAGER_VERSION=0.26.0

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/prometheus/alertmanager
RUN curl -fsSL -o alertmanager.tar.gz "https://github.com/prometheus/alertmanager/archive/v${ALERTMANAGER_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f alertmanager.tar.gz \
    && rm -f alertmanager.tar.gz \
    && make "PREFIX=$GOPATH/bin/alertmanager" build

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:22.04

COPY --from=build /go/bin/alertmanager /bin
COPY --from=build /go/src/github.com/prometheus/alertmanager/LICENSE /
COPY --from=build /go/src/github.com/prometheus/alertmanager/COPYRIGHT.txt /
COPY --from=build /go/src/github.com/prometheus/alertmanager/NOTICE /

USER 10000:10000
EXPOSE 9093
ENTRYPOINT ["/bin/alertmanager"]
