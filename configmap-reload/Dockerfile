# configmap-reload container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.14-bionic AS build

ARG CONFIGMAP_RELOAD_VERSION=0.4.0

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
ENV USER=root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/jimmidyson/configmap-reload
RUN curl -fsSL -o configmap-reload.tar.gz "https://github.com/jimmidyson/configmap-reload/archive/v${CONFIGMAP_RELOAD_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f configmap-reload.tar.gz \
    && rm -f configmap-reload.tar.gz \
    && go install ./

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin/configmap-reload /
COPY --from=build /go/src/github.com/jimmidyson/configmap-reload/LICENSE.txt /

USER 10000:10000

ENTRYPOINT ["/configmap-reload"]
