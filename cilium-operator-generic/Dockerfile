ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.26.7.1_noble@sha256:678dffc420bde6361513f3c63a5c9018f3e8f4ec7480a94f57c2397a81c7da51
ARG UBUNTU_IMAGE=ghcr.io/cybozu/ubuntu:24.04.20260825@sha256:8ff055df3303693be8822fd555cc2cda4f2b16ecc71c6ffdad8c5c74ff054ba7

# Stage1: build
FROM ${GOLANG_IMAGE} as build

COPY TAG /

# LICENSE.all
WORKDIR /go/src/github.com/cybozu-go/
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && BRANCH=v${VERSION}-cybozu \
    && git clone --depth 1 --branch ${BRANCH} https://github.com/cybozu-go/cilium \
    && cd cilium \
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
