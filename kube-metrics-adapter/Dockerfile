# kube-metrics-adapter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.4.2_resolute@sha256:0ddbebb97d9fefa0961dd10fdccedc40a4b805d28f21f21d92faa2f7b071d2f1 AS build
ENV GOPROXY=https://golang.flatt.tech

# https://github.com/zalando-incubator/kube-metrics-adapter/releases/tag/v0.2.10
ARG KMA_VERSION=0.2.10
ARG KMA_COMMIT=2cd43f7be0ddad257ba9e6449dfa80ea9222f8fb

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    curl -fsSL -o kma.tar.gz https://github.com/zalando-incubator/kube-metrics-adapter/archive/${KMA_COMMIT}.tar.gz \
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
