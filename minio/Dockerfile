# minio container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.18-focal AS build

ARG MINIO_SRCREPO=minio/minio
ARG MINIO_VERSION=RELEASE.2022-09-07T22-25-02Z

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /work
RUN git clone --depth=1 -b ${MINIO_VERSION} https://github.com/${MINIO_SRCREPO} minio
WORKDIR /work/minio
RUN CGO_ENABLED=0 go build -trimpath -ldflags "$(go run buildscripts/gen-ldflags.go)" -o minio

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /work/minio/minio /
COPY --from=build /work/minio/LICENSE /

USER 10000:10000

ENTRYPOINT ["/minio"]
