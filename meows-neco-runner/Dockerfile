ARG MEOWS_RUNNER_TAG=2.335.1-ubuntu22.04-meows0.24.1
ARG MEOWS_RUNNER_CHECKSUM=sha256:221f8ff2e85d3e8ee0533a3bebd209ed24ab93889aca6408ef64ac9bbae0b0c1

FROM ghcr.io/cybozu/golang:1.26.4.2_jammy@sha256:64a799f1c3aa5d91c0c73bb629c4a712bd3ea11157e233887092186f0577648c AS builder

WORKDIR /go/src/github.com/fullstorydev/grpcurl
ARG GRPCURL_VERSION=1.9.3
ARG GRPCURL_SHA256=bb555087f279af156159c86d4d3d5dd3f2991129e4cd6b09114e6851a679340d
ADD --checksum=sha256:${GRPCURL_SHA256} \
    "https://github.com/fullstorydev/grpcurl/archive/v${GRPCURL_VERSION}.tar.gz" \
    grpcurl.tar.gz
RUN tar -x -z --strip-components 1 -f grpcurl.tar.gz \
    && GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags "-X \"main.version=${GRPCURL_VERSION}\"" -o /go/bin/grpcurl ./cmd/grpcurl

FROM ghcr.io/cybozu-go/meows-runner:${MEOWS_RUNNER_TAG}@${MEOWS_RUNNER_CHECKSUM}
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
        iperf3 \
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
