# Teleport container image

# Stage1: build teleport
FROM quay.io/cybozu/golang:1.13-bionic AS build

ENV TELEPORT_VERSION=4.2.1
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && \
    apt-get install -y --no-install-recommends zip libpam0g-dev && \
    mkdir -p /go/src/github.com/gravitational
WORKDIR /go/src/github.com/gravitational/
RUN git clone https://github.com/gravitational/teleport.git
WORKDIR /go/src/github.com/gravitational/teleport
RUN git checkout v${TELEPORT_VERSION} && \
    GO111MODULE=off make full

# Stage2: final image
FROM quay.io/cybozu/ubuntu:18.04

RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    dumb-init && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /usr/local/teleport/bin/

COPY --from=build /go/src/github.com/gravitational/teleport/build/tctl /usr/local/teleport/bin/
COPY --from=build /go/src/github.com/gravitational/teleport/build/teleport /usr/local/teleport/bin/
COPY --from=build /go/src/github.com/gravitational/teleport/build/tsh /usr/local/teleport/bin/
COPY --from=build /go/src/github.com/gravitational/teleport/LICENSE /usr/local/teleport/LICENSE
COPY install-tools /usr/local/teleport/install-tools

ENV PATH=/usr/local/teleport/bin/:"$PATH"
RUN mkdir -p /var/lib/teleport && \
    chown 10000:10000 -R /var/lib/teleport
VOLUME /var/lib/teleport

USER 10000:10000

# ssh port
EXPOSE 3022
# proxy ssh
EXPOSE 3023
# proxy tunnel:
EXPOSE 3024
# auth ssh
EXPOSE 3025
# proxy kubernetes
EXPOSE 3026
# proxy web
EXPOSE 3080

ENTRYPOINT ["/usr/bin/dumb-init", "teleport"]
CMD ["start", "-c", "/etc/teleport/teleport.yaml"]
