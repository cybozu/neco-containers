FROM quay.io/cybozu/golang:1.17-focal as build

ARG NERDCTL_VERSION=v0.20.0

ENV SRC_DIR /nerdctl
WORKDIR $SRC_DIR 
RUN git clone https://github.com/containerd/nerdctl.git $SRC_DIR \
        && git checkout refs/tags/$NERDCTL_VERSION

RUN make && make install
 
FROM quay.io/cybozu/ubuntu:20.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /usr/local/bin/nerdctl /usr/local/bin/nerdctl

ENTRYPOINT ["/usr/local/bin/nerdctl", "ipfs", "registry", "serve"]
