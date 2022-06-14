# referring to go-ipfs Dockerfile(https://github.com/ipfs/go-ipfs/blob/master/Dockerfile)
FROM quay.io/cybozu/golang:1.17-focal AS build-idserver
WORKDIR /idserver
COPY idserver/go.mod /idserver/
COPY idserver/go.sum /idserver/
RUN go mod download 

COPY idserver/*.go /idserver/
RUN go build -o /idserver

FROM quay.io/cybozu/golang:1.17-focal AS build-go-ipfs
RUN apt-get update && apt-get install -y \
  fuse \
  pkg-config \
  wget 

ARG GO_IPFS_VERSION=v0.12.2
ENV SRC_DIR /go-ipfs

WORKDIR $SRC_DIR
RUN git clone https://github.com/ipfs/go-ipfs.git $SRC_DIR \
  && git checkout refs/tags/${GO_IPFS_VERSION}

RUN cd $SRC_DIR \
 go mod download 

ARG IPFS_PLUGINS

RUN cd $SRC_DIR \
  && mkdir -p .git/objects \
  && make build GOTAGS=openssl IPFS_PLUGINS=$IPFS_PLUGINS

ENV SUEXEC_VERSION v0.2
RUN set -eux; \
  cd /tmp \
  && git clone https://github.com/ncopa/su-exec.git \
  && cd su-exec \
  && git checkout -q $SUEXEC_VERSION \
  && make su-exec-static 

FROM quay.io/cybozu/ubuntu:20.04 AS build-tini 
ENV TINI_VERSION v0.19.0

RUN apt-get update \
  && apt-get install -y make git cmake

WORKDIR /tini

RUN git clone https://github.com/krallin/tini.git /tini \
  && git checkout refs/tags/$TINI_VERSION \
  && cmake . \
  && make \
  && chmod +x tini

FROM quay.io/cybozu/ubuntu:20.04

RUN apt-get update && apt-get install -y \
  wget \
  netcat 

ENV SRC_DIR /go-ipfs
COPY --from=build-go-ipfs $SRC_DIR/cmd/ipfs/ipfs /usr/local/bin/ipfs
COPY --from=build-go-ipfs $SRC_DIR/bin/container_daemon /usr/local/bin/start_ipfs
COPY --from=build-go-ipfs /tmp/su-exec/su-exec-static /sbin/su-exec
COPY --from=build-tini /tini/tini /sbin/tini
COPY --from=build-go-ipfs /bin/fusermount /usr/local/bin/fusermount
COPY --from=build-go-ipfs /etc/ssl/certs /etc/ssl/certs
COPY --from=build-idserver /idserver/src /idserver

RUN chmod 4755 /usr/local/bin/fusermount
RUN chmod 0755 /usr/local/bin/start_ipfs

COPY --from=build-go-ipfs /usr/lib/*-linux-gnu*/libssl.so* /usr/lib/
COPY --from=build-go-ipfs /usr/lib/*-linux-gnu*/libcrypto.so* /usr/lib/

EXPOSE 4001
EXPOSE 4001/udp
EXPOSE 5001
EXPOSE 8080
EXPOSE 8081

ENV IPFS_PATH /data/ipfs
RUN mkdir -p $IPFS_PATH \
  && adduser --disabled-password --uid 1000 ipfs \
  && chown ipfs:users $IPFS_PATH

RUN mkdir /ipfs /ipns \
  && chown ipfs:users /ipfs /ipns

RUN mkdir /container-init.d \
  && chown ipfs:users /container-init.d

VOLUME $IPFS_PATH

ENV IPFS_LOGGING ""

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/start_ipfs"]

CMD ["daemon", "--migrate=true", "--agent-version-suffix=docker"]

