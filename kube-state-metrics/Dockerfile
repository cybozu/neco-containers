# kube-state-metrics container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG KUBE_STATE_METRICS_VERSION=1.6.0
WORKDIR /go/src/k8s.io/kube-state-metrics
RUN curl -fsSL https://github.com/kubernetes/kube-state-metrics/archive/v${KUBE_STATE_METRICS_VERSION}.tar.gz | \
    tar --strip-components=1 -xzf -

RUN CGO_ENABLED=0 go build -o /go/bin/kube-state-metrics

# Stage2: setup runtime container
FROM scratch
COPY --from=build /go/bin/kube-state-metrics /usr/local/kube-state-metrics/bin/kube-state-metrics
COPY --from=build /go/src/k8s.io/kube-state-metrics/LICENSE /usr/local/kube-state-metrics/LICENSE
ENV PATH=/usr/local/kube-state-metrics/bin:"$PATH"

USER 10000:10000
EXPOSE 8080 8081
ENTRYPOINT ["kube-state-metrics", "--port=8080", "--telemetry-port=8081"]
