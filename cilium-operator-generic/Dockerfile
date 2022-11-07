ARG BASE_IMAGE=scratch
ARG GOLANG_IMAGE=quay.io/cybozu/golang:1.17-focal
ARG UBUNTU_IMAGE=quay.io/cybozu/ubuntu:20.04

# Stage1: build
FROM ${GOLANG_IMAGE} as build

COPY TAG /
COPY 17930.patch /tmp/
COPY 18454.patch /tmp/
COPY 20092.patch /tmp/
COPY 20176.patch /tmp/

# LICENSE.all
WORKDIR /go/src/github.com/cilium/cilium
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL "https://github.com/cilium/cilium/archive/v${VERSION}.tar.gz" | \
      tar xzf - --strip-components 1 \
    && patch -p1 --no-backup-if-mismatch < /tmp/17930.patch \
    && patch -p1 --no-backup-if-mismatch < /tmp/18454.patch \
    && patch -p1 --no-backup-if-mismatch < /tmp/20092.patch \
    && patch -p1 --no-backup-if-mismatch < /tmp/20176.patch \
    && make licenses-all \
    && apt-get update \
    && apt-get install -y --no-install-recommends binutils-aarch64-linux-gnu \
    && images/runtime/build-gops.sh \
    && mkdir -p /etc/gops

# operator-generic
WORKDIR /go/src/github.com/cilium/cilium/operator
RUN make cilium-operator-generic

# Stage2: runtime
FROM ${BASE_IMAGE}
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/linux/amd64/bin/gops /bin/gops
COPY --from=build --chown=10000:10000 /etc/gops /etc/gops
COPY --from=build /go/src/github.com/cilium/cilium/LICENSE.all /LICENSE
COPY --from=build /go/src/github.com/cilium/cilium/operator/cilium-operator-generic /usr/bin/cilium-operator-generic

USER 10000:10000

WORKDIR /
ENV GOPS_CONFIG_DIR=/etc/gops

ENTRYPOINT ["/usr/bin/cilium-operator-generic"]
