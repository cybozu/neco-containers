# kube-metrics-adapter container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.16-focal AS build

ARG KMA_VERSION=0.1.10

RUN curl -fsSL -o kma.tar.gz https://github.com/zalando-incubator/kube-metrics-adapter/archive/v${KMA_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f kma.tar.gz \
    && rm -f kma.tar.gz \
    && make build.linux VERSION=${KMA_VERSION}

# Stage2: setup runtime container
FROM scratch

COPY --from=build /work/build/linux/kube-metrics-adapter /
COPY --from=build /work/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/kube-metrics-adapter"]
