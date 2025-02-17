# kube-metrics-adapter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.23-noble AS build

ARG KMA_VERSION=0.2.3

RUN curl -fsSL -o kma.tar.gz https://github.com/zalando-incubator/kube-metrics-adapter/archive/v${KMA_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f kma.tar.gz \
    && rm -f kma.tar.gz \
    && make build.linux VERSION=${KMA_VERSION}

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/build/linux/kube-metrics-adapter /
COPY --from=build /work/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/kube-metrics-adapter"]
