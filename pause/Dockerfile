# pause container

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:26.04.20261117@sha256:d1fe8476ab613ca5a8320ba5c79cd5a50854832e89831014a3732a2816233e8f AS build

# https://github.com/kubernetes/kubernetes/releases/tag/v1.36.6
ARG K8S_COMMIT=1b2171b110bb2a2542e1322035c6583ec86eb976
ARG PAUSE_VERSION=3.10.2

RUN mkdir /work

WORKDIR /work

RUN curl -sSLf -O https://raw.githubusercontent.com/kubernetes/kubernetes/${K8S_COMMIT}/build/pause/linux/pause.c \
    && gcc -Os -Wall -Werror -static -DVERSION=v${PAUSE_VERSION} -o pause pause.c \
    && strip pause

RUN curl -sSLf -O https://github.com/kubernetes/kubernetes/raw/${K8S_COMMIT}/LICENSE


# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/pause /pause
COPY --from=build /work/LICENSE /LICENSE

USER 65535:65535

ENTRYPOINT ["/pause"]
