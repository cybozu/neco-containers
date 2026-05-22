# kube-metrics-adapter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG KMA_VERSION=0.2.9
ARG KMA_COMMIT=77733fb03d3fb379abebdca8c7d8a9716392c0a0

RUN curl -fsSL -o kma.tar.gz https://github.com/zalando-incubator/kube-metrics-adapter/archive/${KMA_COMMIT}.tar.gz \
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
