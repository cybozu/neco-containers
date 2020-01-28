# Build stage
FROM quay.io/cybozu/golang:1.13-bionic AS build

ARG METRICS_SERVER_VERSION=0.3.6
ARG CGO_ENABLED=0
WORKDIR /go/src/github.com/kubernetes-incubator/metrics-server
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/kubernetes-sigs/metrics-server/archive/v${METRICS_SERVER_VERSION}.tar.gz | \
    tar --strip-components=1 -xzf - && \
    make all

# Runtime stage
FROM scratch
COPY --from=build /go/src/github.com/kubernetes-incubator/metrics-server/_output/amd64/metrics-server /usr/local/metrics-server/bin/metrics-server
COPY --from=build /go/src/github.com/kubernetes-incubator/metrics-server/LICENSE /usr/local/metrics-server/bin/LICENSE
ENV PATH=/usr/local/metrics-server/bin:"$PATH"

USER 65534:65534
EXPOSE 4443
ENTRYPOINT ["metrics-server"]
