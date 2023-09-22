ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=quay.io/cybozu/golang:1.20-jammy
ARG UBUNTU_IMAGE=quay.io/cybozu/ubuntu:22.04

# Stage1: build
FROM ${GOLANG_IMAGE} as build

COPY TAG /

# LICENSE.all
WORKDIR /go/src/github.com/cybozu-go/
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    # Since we use the fork and patched repogitory, we need to specify the branch name
    && BRANCH=v${VERSION}-lb-dsr-patch \
    && echo ${BRANCH} \
    && git clone --depth 1 --branch ${BRANCH} https://github.com/cybozu-go/cilium \
    && cd cilium \
    && make licenses-all \
    && apt-get update \
    && apt-get install -y --no-install-recommends binutils-aarch64-linux-gnu \
    && images/runtime/build-gops.sh

# hubble-relay
WORKDIR /go/src/github.com/cybozu-go/cilium/hubble-relay
RUN make

# Stage2: runtime
FROM ${BASE_IMAGE}
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/linux/amd64/bin/gops /bin/gops
COPY --from=build /go/src/github.com/cybozu-go/cilium/LICENSE.all /LICENSE
COPY --from=build /go/src/github.com/cybozu-go/cilium/hubble-relay/hubble-relay /usr/bin/hubble-relay

WORKDIR /
ENV GOPS_CONFIG_DIR=/

ENTRYPOINT ["/usr/bin/hubble-relay"]
CMD ["serve"]
