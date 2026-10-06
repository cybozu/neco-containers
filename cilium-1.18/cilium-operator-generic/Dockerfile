ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.26.7.1_noble@sha256:678dffc420bde6361513f3c63a5c9018f3e8f4ec7480a94f57c2397a81c7da51
ARG UBUNTU_IMAGE=ghcr.io/cybozu/ubuntu:24.04.20260825@sha256:8ff055df3303693be8822fd555cc2cda4f2b16ecc71c6ffdad8c5c74ff054ba7

# Stage1: build
FROM ${GOLANG_IMAGE} as build

# LICENSE.all
WORKDIR /go/src/github.com/cybozu-go/
ENV CILIUM_REPOSITORY=https://github.com/cybozu-go/cilium
ENV CILIUM_SHA=5dc1ce3e109526bc170351800379fdc9903c18be
RUN mkdir cilium \
    && cd cilium \
    && git init 2>/dev/null \
    && git remote add origin ${CILIUM_REPOSITORY} \
    && git fetch --depth 1 origin ${CILIUM_SHA} \
    && git checkout FETCH_HEAD 2>/dev/null \
    && make licenses-all \
    && apt-get update \
    && apt-get install -y --no-install-recommends binutils-aarch64-linux-gnu \
    && images/runtime/build-gops.sh \
    && mkdir -p /etc/gops

# operator-generic
WORKDIR /go/src/github.com/cybozu-go/cilium/operator
RUN make cilium-operator-generic

# Stage2: runtime
FROM ${BASE_IMAGE}
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/linux/amd64/bin/gops /bin/gops
COPY --from=build --chown=10000:10000 /etc/gops /etc/gops
COPY --from=build /go/src/github.com/cybozu-go/cilium/LICENSE.all /LICENSE
COPY --from=build /go/src/github.com/cybozu-go/cilium/operator/cilium-operator-generic /usr/bin/cilium-operator-generic

USER 10000:10000

WORKDIR /
ENV GOPS_CONFIG_DIR=/etc/gops

ENTRYPOINT ["/usr/bin/cilium-operator-generic"]
