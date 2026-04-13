ARG UBUNTU_IMAGE=ghcr.io/cybozu/ubuntu:24.04
ARG GOLANG_IMAGE=ghcr.io/cybozu/golang:1.24-noble
ARG CILIUM_SHA=011fab6e7b798edd17e0337307dc034ed4916653

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
