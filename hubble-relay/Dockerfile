ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=quay.io/cybozu/golang:1.20-jammy
ARG UBUNTU_IMAGE=quay.io/cybozu/ubuntu:22.04

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

# hubble-relay
WORKDIR /go/src/github.com/cilium/cilium/hubble-relay
RUN make

# Stage2: runtime
FROM ${BASE_IMAGE}
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/linux/amd64/bin/gops /bin/gops
COPY --from=build /go/src/github.com/cilium/cilium/LICENSE.all /LICENSE
COPY --from=build /go/src/github.com/cilium/cilium/hubble-relay/hubble-relay /usr/bin/hubble-relay

WORKDIR /
ENV GOPS_CONFIG_DIR=/

ENTRYPOINT ["/usr/bin/hubble-relay"]
CMD ["serve"]
