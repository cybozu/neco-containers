# tcp-keepalive container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -ldflags="-w -s" .

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04.20260716@sha256:d53016ed374d33c73c32e9b5674994797f59df41bc0f8252034325a2436bf815
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt update && \
	apt install -y iproute2 tcpdump iputils-ping iptables net-tools inetutils-traceroute dnsutils

COPY --from=build /go/bin /usr/bin
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/usr/bin/tcp-keepalive"]
