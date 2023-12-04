# pause container

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:22.04 AS build

ARG K8S_VERSION=1.27.8
ARG PAUSE_VERSION=3.9

RUN mkdir /work

WORKDIR /work

RUN curl -sSLf -O https://raw.githubusercontent.com/kubernetes/kubernetes/v${K8S_VERSION}/build/pause/linux/pause.c \
    && gcc -Os -Wall -Werror -static -DVERSION=v${PAUSE_VERSION} -o pause pause.c \
    && strip pause

RUN curl -sSLf -O https://github.com/kubernetes/kubernetes/raw/v${K8S_VERSION}/LICENSE


# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/pause /pause
COPY --from=build /work/LICENSE /LICENSE

USER 65535:65535

ENTRYPOINT ["/pause"]
