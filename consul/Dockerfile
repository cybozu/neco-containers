# consul container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.16-focal AS build

ARG CONSUL_VERSION=1.9.4

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN git clone --depth=1 -b v${CONSUL_VERSION} https://github.com/hashicorp/consul /work/consul

WORKDIR /work/consul
RUN CGO_ENABLED=0 go install -ldflags="-w -s"

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /go/bin/consul /bin/consul
RUN mkdir -p /consul/data && \
    mkdir -p /consul/config && \
    chown -R 10000:10000 /consul

USER 10000:10000
VOLUME /consul/data
EXPOSE 8300
EXPOSE 8301 8301/udp 8302 8302/udp
EXPOSE 8500 8600 8600/udp
ENTRYPOINT ["/bin/consul"]
CMD ["agent", "-dev", "-client", "0.0.0.0"]
