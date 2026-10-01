ARG UBUNTU_IMAGE=ghcr.io/cybozu/ubuntu:24.04.20260825@sha256:8ff055df3303693be8822fd555cc2cda4f2b16ecc71c6ffdad8c5c74ff054ba7
ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.26.7.1_noble@sha256:678dffc420bde6361513f3c63a5c9018f3e8f4ec7480a94f57c2397a81c7da51

# https://github.com/cilium/cilium/releases/tag/v1.18.6
ARG CILIUM_SHA=95896696119f33a0f5d21019182192fcdab3600f

# Stage1: build
FROM ${GOLANG_IMAGE} AS build
ARG CILIUM_SHA

COPY TAG /

# LICENSE.all
WORKDIR /go/src/github.com/cilium/
RUN set -eux; \
    VERSION=$(cut -d '.' -f 1-3 < /TAG ); \
    CILIUM_TAG=v${VERSION}; \
    git init cilium; cd cilium; \
    git remote add origin https://github.com/cilium/cilium; \
    git fetch --depth 1 origin "refs/tags/${CILIUM_TAG}"; \
    test "$(git rev-parse 'FETCH_HEAD^{commit}')" = "${CILIUM_SHA}"; \
    git checkout --detach "${CILIUM_SHA}"; \
    git log -n1 --oneline; \
    make licenses-all; \
    apt-get update; \
    apt-get install -y --no-install-recommends binutils-aarch64-linux-gnu; \
    images/runtime/build-gops.sh

# hubble-relay
WORKDIR /go/src/github.com/cilium/cilium/hubble-relay
RUN make

# Stage2: runtime
FROM ${UBUNTU_IMAGE}
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

# Add nonroot user for the neco environment
RUN useradd nonroot -u 10000 -U -m

COPY --from=build /out/linux/amd64/bin/gops /bin/gops
COPY --from=build /go/src/github.com/cilium/cilium/LICENSE.all /LICENSE
COPY --from=build /go/src/github.com/cilium/cilium/hubble-relay/hubble-relay /usr/bin/hubble-relay

# use uid:gid for the nonroot user for compatibility with runAsNonRoot
USER 10000:10000

ENV HOME=/home/nonroot
ENTRYPOINT ["/usr/bin/hubble-relay"]
CMD ["serve"]
