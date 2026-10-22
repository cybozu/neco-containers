# testhttpd container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build
ENV GOPROXY=https://golang.flatt.tech
COPY src /work/src
WORKDIR /work/src
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    CGO_ENABLED=0 go install -ldflags="-w -s" ./testhttpd

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04.20261021@sha256:7bc89f8e6e4dfe609ca671526e3bbee91753422b8a766d90cef4de8b7c73a3b9
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"
RUN apt update && \
	apt install -y iproute2 tcpdump iputils-ping inetutils-traceroute && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /go/bin /
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/testhttpd", "-listen", ":8000"]
