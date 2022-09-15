# mc (minio client) container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.18-focal AS build

ARG MC_SRCREPO=minio/mc
ARG MC_VERSION=RELEASE.2022-08-28T20-08-11Z

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /work
RUN git clone --depth=1 -b ${MC_VERSION} https://github.com/${MC_SRCREPO} mc
WORKDIR /work/mc
RUN CGO_ENABLED=0 go build -trimpath -ldflags "$(go run buildscripts/gen-ldflags.go)" -o mc

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /work/mc/mc /
COPY --from=build /work/mc/LICENSE /

USER 10000:10000

ENTRYPOINT ["/mc"]
