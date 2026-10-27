# syntax=docker/dockerfile:1
ARG MEOWS_RUNNER_TAG=2.338.1-ubuntu24.04-meows0.29.1
ARG MEOWS_RUNNER_CHECKSUM=sha256:f55776fd3a8c2fb65214da44de25e6cd9d59a351a50c1315e51c268fcc1a9adc
ARG AWSCLI_VERSION=2.36.40

FROM scratch AS awscli-download-amd64
ARG AWSCLI_VERSION
ADD --checksum=sha256:a904e314340ad4c4b62beb50e9259d4becb2acf3dffc86c96c553c834d7fdf7a \
    https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_VERSION}.zip /awscliv2.zip

FROM scratch AS awscli-download-arm64
ARG AWSCLI_VERSION
ADD --checksum=sha256:f26e08d0d64d9f35f304a9df94d63cbd11ffe8881c8b4bf60be934cf9e4d13a9 \
    https://awscli.amazonaws.com/awscli-exe-linux-aarch64-${AWSCLI_VERSION}.zip /awscliv2.zip

FROM awscli-download-${TARGETARCH} AS awscli-download

FROM ghcr.io/cybozu/golang:1.26.10.2_noble@sha256:8d900a9ee78c02dbe4e2d2e23800c4d21714dfaf8c679935afa668ca0993a0e4 AS builder
ENV GOPROXY=https://golang.flatt.tech

WORKDIR /go/src/github.com/fullstorydev/grpcurl
ARG GRPCURL_VERSION=1.9.3
ARG GRPCURL_SHA256=bb555087f279af156159c86d4d3d5dd3f2991129e4cd6b09114e6851a679340d
ADD --checksum=sha256:${GRPCURL_SHA256} \
    "https://github.com/fullstorydev/grpcurl/archive/v${GRPCURL_VERSION}.tar.gz" \
    grpcurl.tar.gz
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    tar -x -z --strip-components 1 -f grpcurl.tar.gz \
    && GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags "-X \"main.version=${GRPCURL_VERSION}\"" -o /go/bin/grpcurl ./cmd/grpcurl

FROM ghcr.io/cybozu-go/meows-runner:${MEOWS_RUNNER_TAG}@${MEOWS_RUNNER_CHECKSUM}
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

USER root

# Disable gh telemetry: https://cli.github.com/telemetry
ENV GH_TELEMETRY=false
ENV DO_NOT_TRACK=true

RUN curl -sSLf https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list

RUN COMMON_PACKAGES=" \
        apache2-utils \
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
        python3-requests \
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
        unzip \
        vim \
    " \
    CUSTOM_PACKAGES=" \
        build-essential \
        gh \
        git \
    " \
    && apt-get update && apt-get install -y --no-install-recommends $COMMON_PACKAGES $CUSTOM_PACKAGES \
    && rm -rf /var/lib/apt/lists/* \
    && setcap -r /usr/bin/ping

WORKDIR /work
RUN --mount=type=bind,from=awscli-download,source=/awscliv2.zip,target=/work/awscliv2.zip \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf ./aws

COPY install-necocli /usr/local/bin
COPY --from=builder /go/bin/grpcurl /usr/local/bin
COPY --from=builder /go/src/github.com/fullstorydev/grpcurl/LICENSE /usr/local/grpcurl/LICENSE

USER 10000
