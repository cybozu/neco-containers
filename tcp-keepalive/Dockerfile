# tcp-keepalive container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build
COPY src /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./tcp-keepalive

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt update && \
	apt install -y iproute2 tcpdump iputils-ping iptables net-tools inetutils-traceroute dnsutils

COPY --from=build /go/bin /usr/bin
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/usr/bin/tcp-keepalive"]
