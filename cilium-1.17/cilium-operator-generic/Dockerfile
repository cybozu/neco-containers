ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.24-noble
ARG UBUNTU_IMAGE=ghcr.io/cybozu/ubuntu:24.04

# Stage1: build
FROM ${GOLANG_IMAGE} as build

# LICENSE.all
WORKDIR /go/src/github.com/cybozu-go/
ENV CILIUM_REPOSITORY=https://github.com/cybozu-go/cilium
ENV CILIUM_SHA=3e401f859afa255e6156f0e1d45adf086a80285b
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
