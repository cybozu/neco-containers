FROM quay.io/cybozu/golang:1.17-focal AS build-idserver
WORKDIR /idserver
COPY idserver/go.mod /idserver/
COPY idserver/go.sum /idserver/
RUN go mod download 

COPY idserver/*.go /idserver/
RUN go build -o /idserver

FROM quay.io/cybozu/golang:1.17-focal AS build-go-ipfs

# Install deps
RUN apt-get update && apt-get install -y \
  fuse \
  pkg-config \
  wget 

ARG GO_IPFS_VERSION=v0.12.2
ENV SRC_DIR /go-ipfs

WORKDIR $SRC_DIR
RUN git clone https://github.com/ipfs/go-ipfs.git $SRC_DIR \
  && git checkout refs/tags/${GO_IPFS_VERSION}

# Download packages first so they can be cached.
RUN cd $SRC_DIR \
 go mod download 

# Preload an in-tree but disabled-by-default plugin by adding it to the IPFS_PLUGINS variable
# e.g. docker build --build-arg IPFS_PLUGINS="foo bar baz"
ARG IPFS_PLUGINS

# Build the thing.
# Also: fix getting HEAD commit hash via git rev-parse.
RUN cd $SRC_DIR \
  && mkdir -p .git/objects \
  && make build GOTAGS=openssl IPFS_PLUGINS=$IPFS_PLUGINS

# Get su-exec, a very minimal tool for dropping privileges,
# and tini, a very minimal init daemon for containers
ENV SUEXEC_VERSION v0.2
ENV TINI_VERSION v0.19.0
RUN set -eux; \
    dpkgArch="$(dpkg --print-architecture)"; \
    case "${dpkgArch##*-}" in \
        "amd64" | "armhf" | "arm64") tiniArch="tini-static-$dpkgArch" ;;\
        *) echo >&2 "unsupported architecture: ${dpkgArch}"; exit 1 ;; \
    esac; \
  cd /tmp \
  && git clone https://github.com/ncopa/su-exec.git \
  && cd su-exec \
  && git checkout -q $SUEXEC_VERSION \
  && make su-exec-static \
  && cd /tmp \
  && wget -q -O tini https://github.com/krallin/tini/releases/download/$TINI_VERSION/$tiniArch \
  && chmod +x tini

# Now comes the actual target image, which aims to be as small as possible.
FROM quay.io/cybozu/ubuntu:20.04

RUN apt-get update && apt-get install -y \
  wget \
  netcat 

# Get the ipfs binary, entrypoint script, and TLS CAs from the build container.
ENV SRC_DIR /go-ipfs
COPY --from=build-go-ipfs $SRC_DIR/cmd/ipfs/ipfs /usr/local/bin/ipfs
COPY --from=build-go-ipfs $SRC_DIR/bin/container_daemon /usr/local/bin/start_ipfs
COPY --from=build-go-ipfs /tmp/su-exec/su-exec-static /sbin/su-exec
COPY --from=build-go-ipfs /tmp/tini /sbin/tini
COPY --from=build-go-ipfs /bin/fusermount /usr/local/bin/fusermount
COPY --from=build-go-ipfs /etc/ssl/certs /etc/ssl/certs

COPY --from=build-idserver /idserver/src /idserver

# Add suid bit on fusermount so it will run properly
RUN chmod 4755 /usr/local/bin/fusermount

# Fix permissions on start_ipfs (ignore the build machine's permissions)
RUN chmod 0755 /usr/local/bin/start_ipfs

# Copy over SSL libraries.
COPY --from=build-go-ipfs /usr/lib/*-linux-gnu*/libssl.so* /usr/lib/
COPY --from=build-go-ipfs /usr/lib/*-linux-gnu*/libcrypto.so* /usr/lib/

# Swarm TCP; should be exposed to the public
EXPOSE 4001
# Swarm UDP; should be exposed to the public
EXPOSE 4001/udp
# Daemon API; must not be exposed publicly but to client services under you control
EXPOSE 5001
# Web Gateway; can be exposed publicly with a proxy, e.g. as https://ipfs.example.org
EXPOSE 8080
# Swarm Websockets; must be exposed publicly when the node is listening using the websocket transport (/ipX/.../tcp/8081/ws).
EXPOSE 8081

# Create the fs-repo directory and switch to a non-privileged user.
ENV IPFS_PATH /data/ipfs
RUN mkdir -p $IPFS_PATH \
  && adduser --disabled-password --uid 1000 ipfs \
  && chown ipfs:users $IPFS_PATH

# Create mount points for `ipfs mount` command
RUN mkdir /ipfs /ipns \
  && chown ipfs:users /ipfs /ipns

# Create the init scripts directory
RUN mkdir /container-init.d \
  && chown ipfs:users /container-init.d

# Expose the fs-repo as a volume.
# start_ipfs initializes an fs-repo if none is mounted.
# Important this happens after the USER directive so permissions are correct.
VOLUME $IPFS_PATH

# The default logging level
ENV IPFS_LOGGING ""

# This just makes sure that:
# 1. There's an fs-repo, and initializes one if there isn't.
# 2. The API and Gateway are accessible from outside the container.
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/start_ipfs"]

# Execute the daemon subcommand by default
CMD ["daemon", "--migrate=true", "--agent-version-suffix=docker"]
