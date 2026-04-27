# pause container

# Stage1: build from source
FROM ghcr.io/cybozu/ubuntu-dev:24.04 AS build

# v1.35.3
ARG K8S_COMMIT=6c1cd99aef09161ddb07b8ade6c9564e9b9bfe27
ARG PAUSE_VERSION=3.10.1

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
