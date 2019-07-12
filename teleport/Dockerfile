# Teleport container image

# Stage1: build teleport
FROM quay.io/cybozu/golang:1.12-bionic AS build

ENV TELEPORT_VERSION=4.0.2
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && \
    apt-get install -y --no-install-recommends zip && \ 
    mkdir -p /go/src/github.com/gravitational
WORKDIR /go/src/github.com/gravitational/
RUN git clone https://github.com/gravitational/teleport.git
WORKDIR /go/src/github.com/gravitational/teleport
RUN git checkout v${TELEPORT_VERSION} && \
    GO111MODULE=off make full

# Stage2: final image
FROM quay.io/cybozu/ubuntu:18.04

RUN apt-get update && apt-get upgrade -y && \
    apt-get install --no-install-recommends -y \
    dumb-init \
    ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /usr/local/teleport/bin/teleport

COPY --from=build /go/src/github.com/gravitational/teleport/build/tctl /usr/local/teleport/bin/teleport/
COPY --from=build /go/src/github.com/gravitational/teleport/build/teleport /usr/local/teleport/bin/teleport/
COPY --from=build /go/src/github.com/gravitational/teleport/build/tsh /usr/local/teleport/bin/teleport/
ENV PATH=/usr/local/teleport/bin/teleport:"$PATH"
RUN mkdir -p /var/lib/teleport && \
    chown 10000:10000 -R /var/lib/teleport

USER 10000:10000

ENTRYPOINT ["/usr/bin/dumb-init", "teleport", "start", "-c", "/etc/teleport/teleport.yaml"]
