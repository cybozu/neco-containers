FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG K8S_METACOLLECTOR_COMMIT=d0d27fb991d8cdfad29de28ffdb68cab75cc13a9 # v0.1.2

WORKDIR /work/k8s-metacollector
RUN curl -fsSL -o /tmp/k8s-metacollector-${K8S_METACOLLECTOR_COMMIT}.tar.gz https://github.com/falcosecurity/k8s-metacollector/archive/${K8S_METACOLLECTOR_COMMIT}.tar.gz && \
    tar -xzf /tmp/k8s-metacollector-${K8S_METACOLLECTOR_COMMIT}.tar.gz --strip-components=1 && \
    rm /tmp/k8s-metacollector-${K8S_METACOLLECTOR_COMMIT}.tar.gz && \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o /usr/local/bin/meta-collector main.go

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /usr/local/bin/meta-collector /usr/bin/meta-collector
COPY --from=build /work/k8s-metacollector/LICENSE /usr/local/falco-k8s-metacollector/LICENSE

USER 65532:65532

ENTRYPOINT ["/usr/bin/meta-collector"]
