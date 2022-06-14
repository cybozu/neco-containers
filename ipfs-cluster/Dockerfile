# referring to ipfs-cluster Dockerfile (https://github.com/ipfs/ipfs-cluster/blob/master/Dockerfile)
FROM quay.io/cybozu/golang:1.18-focal AS builder

# This dockerfile builds and runs ipfs-cluster-service.
ENV GOPATH      /go
ENV SRC_PATH    $GOPATH/src/github.com/ipfs/ipfs-cluster
ENV GOPROXY     https://proxy.golang.org

ENV SUEXEC_VERSION v0.2
ENV TINI_VERSION v0.19.0
ENV IPFS_CLUSTER_VERSION v1.0.1

RUN apt-get update && apt-get install -y \
    cmake

RUN set -eux; \
  cd /tmp \
  && git clone https://github.com/ncopa/su-exec.git \
  && cd su-exec \
  && git checkout -q $SUEXEC_VERSION \
  && make su-exec-static 

RUN cd /tmp \
  && git clone https://github.com/krallin/tini.git  \
  && cd tini \
  && git checkout refs/tags/$TINI_VERSION \
  && cmake . \
  && make \
  && chmod +x tini

WORKDIR $SRC_PATH
RUN git clone https://github.com/ipfs/ipfs-cluster.git $SRC_PATH \
    && git checkout refs/tags/${IPFS_CLUSTER_VERSION}
RUN go mod download

COPY --chown=1000:users . $SRC_PATH
RUN make install


#------------------------------------------------------
FROM quay.io/cybozu/ubuntu:20.04 

ENV GOPATH                 /go
ENV SRC_PATH               /go/src/github.com/ipfs/ipfs-cluster
ENV IPFS_CLUSTER_PATH      /data/ipfs-cluster
ENV IPFS_CLUSTER_CONSENSUS crdt
ENV IPFS_CLUSTER_DATASTORE leveldb

RUN apt-get update && apt-get install -y \
    netcat

EXPOSE 9094
EXPOSE 9095
EXPOSE 9096

COPY --from=builder $GOPATH/bin/ipfs-cluster-service /usr/local/bin/ipfs-cluster-service
COPY --from=builder $GOPATH/bin/ipfs-cluster-ctl /usr/local/bin/ipfs-cluster-ctl
COPY --from=builder $GOPATH/bin/ipfs-cluster-follow /usr/local/bin/ipfs-cluster-follow
COPY --from=builder $SRC_PATH/docker/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=builder /tmp/su-exec/su-exec-static /sbin/su-exec
COPY --from=builder /tmp/tini/tini /sbin/tini
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

RUN mkdir -p $IPFS_CLUSTER_PATH && \
    adduser --disabled-password --uid 1000 ipfs && \
    chown ipfs:users $IPFS_CLUSTER_PATH

VOLUME $IPFS_CLUSTER_PATH
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/entrypoint.sh"]

# Defaults for ipfs-cluster-service go here
CMD ["daemon"]
