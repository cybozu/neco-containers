# tcp-keepalive container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.3.2_resolute@sha256:2e42484cb145c45c51065e58a62420b3afa2164ed303062ceb2f7b82e978df20 AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -ldflags="-w -s" .

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:26.04.20261021@sha256:4d40515de878b3276fa0b0fd6ee4818f7dc729d579d3a211dae0763bfad4ff02
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt update && \
	apt install -y iproute2 tcpdump iputils-ping iptables net-tools inetutils-traceroute dnsutils

COPY --from=build /go/bin /usr/bin
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/usr/bin/tcp-keepalive"]
