ARG MEOWS_VERSION=0.20.1

FROM ghcr.io/cybozu/golang:1.23-jammy AS builder

WORKDIR /go/src/github.com/fullstorydev/grpcurl
ARG GRPCURL_VERSION=1.9.2
RUN curl -fsSL -o grpcurl.tar.gz "https://github.com/fullstorydev/grpcurl/archive/v${GRPCURL_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f grpcurl.tar.gz \
    && rm -f grpcurl.tar.gz \
    && GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags "-X \"main.version=${GRPCURL_VERSION}\"" -o /go/bin/grpcurl ./cmd/grpcurl

FROM ghcr.io/cybozu-go/meows-runner:${MEOWS_VERSION}
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

USER root

# Install debug tools.
# https://github.com/cybozu/ubuntu-base/blob/main/22.04/ubuntu-debug/Dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends \
        apache2-utils \
        awscli \
        bash-completion \
        bind9-dnsutils \
        binutils \
        bsdmainutils \
        conntrack \
        expect \
        file \
        gdb \
        iproute2 \
        iputils-ping \
        jq \
        less \
        lsb-release \
        lv \
        moreutils \
        mysql-client \
        net-tools \
        openssh-client \
        perl-modules \
        rclone \
        rsync \
        s3cmd \
        smartmontools \
        socat \
        sqlite3 \
        strace \
        sysstat \
        tcpdump \
        telnet \
        traceroute \
        vim \
    && rm -rf /var/lib/apt/lists/* \
    && setcap -r /usr/bin/ping

COPY install-necocli /usr/local/bin
COPY --from=builder /go/bin/grpcurl /usr/local/bin
COPY --from=builder /go/src/github.com/fullstorydev/grpcurl/LICENSE /usr/local/grpcurl/LICENSE

USER 10000
