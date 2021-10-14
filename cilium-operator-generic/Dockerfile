ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=quay.io/cybozu/golang:1.17-focal
ARG UBUNTU_IMAGE=quay.io/cybozu/ubuntu:20.04

# Stage1: build
FROM ${GOLANG_IMAGE} as build

COPY TAG /

# LICENSE.all
WORKDIR /go/src/github.com/cilium/cilium
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL "https://github.com/cilium/cilium/archive/v${VERSION}.tar.gz" | \
      tar xzf - --strip-components 1 \
    && make licenses-all \
    && apt-get update \
    && apt-get install -y --no-install-recommends binutils-aarch64-linux-gnu \
    && images/runtime/build-gops.sh

# operator-generic
WORKDIR /go/src/github.com/cilium/cilium/operator
RUN make cilium-operator-generic

# Stage2: runtime
FROM ${BASE_IMAGE}
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/linux/amd64/bin/gops /bin/gops
COPY --from=build /go/src/github.com/cilium/cilium/LICENSE.all /LICENSE
COPY --from=build /go/src/github.com/cilium/cilium/operator/cilium-operator-generic /

USER 10000:10000

WORKDIR /
ENV GOPS_CONFIG_DIR=/

ENTRYPOINT ["/cilium-operator-generic"]
