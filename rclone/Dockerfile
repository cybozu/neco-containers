# rclone container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.18-focal AS build

ARG RCLONE_SRCREPO=rclone/rclone
# we need --s3-version-at option. we cannot use v1.59.x
# master branch on 2022-09-22
ARG RCLONE_REVISION=7e547822d667e644576d250655fa3013a7159076

# Workaround https://github.com/ksonnet/ksonnet/issues/298#issuecomment-360531855
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /work
#RUN git clone --depth=1 -b ${RCLONE_REVISION} https://github.com/${RCLONE_SRCREPO} rclone
RUN git clone https://github.com/${RCLONE_SRCREPO} rclone && \
    cd rclone && git checkout ${RCLONE_REVISION} # in case of revision hash, we cannot use -b option
WORKDIR /work/rclone
RUN make rclone

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /work/rclone/rclone /
COPY --from=build /work/rclone/COPYING /

USER 10000:10000

ENTRYPOINT ["/rclone"]
