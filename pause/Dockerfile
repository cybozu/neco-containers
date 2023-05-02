# pause container

# Stage1: build from source
FROM quay.io/cybozu/ubuntu-dev:22.04 AS build

ARG K8S_VERSION=1.25.9
ARG PAUSE_VERSION=3.8

RUN mkdir /work

WORKDIR /work

RUN curl -sSLf -O https://raw.githubusercontent.com/kubernetes/kubernetes/v${K8S_VERSION}/build/pause/linux/pause.c \
    && gcc -Os -Wall -Werror -static -DVERSION=v${PAUSE_VERSION} -o pause pause.c \
    && strip pause

RUN curl -sSLf -O https://github.com/kubernetes/kubernetes/raw/v${K8S_VERSION}/LICENSE


# Stage2: setup runtime container
FROM scratch

COPY --from=build /work/pause /pause
COPY --from=build /work/LICENSE /LICENSE

USER 65535:65535

ENTRYPOINT ["/pause"]
