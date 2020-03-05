# pause container

# Stage1: build from source
FROM quay.io/cybozu/ubuntu-dev:18.04 AS build

ARG PAUSE_VERSION=3.2

RUN mkdir /work

WORKDIR /work

RUN curl -sSLf -O https://github.com/kubernetes/kubernetes/raw/master/build/pause/pause.c \
    && gcc -Os -Wall -Werror -static -DVERSION=v${PAUSE_VERSION} -o pause pause.c \
    && strip pause

RUN curl -sSLf -O https://github.com/kubernetes/kubernetes/raw/master/LICENSE


# Stage2: setup runtime container
FROM scratch

COPY --from=build /work/pause /pause
COPY --from=build /work/LICENSE /LICENSE

ENTRYPOINT ["/pause"]
