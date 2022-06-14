FROM quay.io/cybozu/golang:1.17-focal 

ARG NERDCTL_VERSION=v0.20.0

ENV SRC_DIR /nerdctl
WORKDIR $SRC_DIR 
RUN git clone https://github.com/containerd/nerdctl.git $SRC_DIR \
        && git checkout refs/tags/$NERDCTL_VERSION

RUN make && make install
 
ENTRYPOINT ["/usr/local/bin/nerdctl", "ipfs", "registry", "serve"]
