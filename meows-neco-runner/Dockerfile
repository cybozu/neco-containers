ARG MEOWS_VERSION=0.8.0

FROM quay.io/cybozu/meows-runner:${MEOWS_VERSION}

USER root

# Install debug tools.
# https://github.com/cybozu/ubuntu-base/blob/main/20.04/ubuntu-debug/Dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends \
        awscli \
        bind9-dnsutils \
        binutils \
        expect \
        file \
        gdb \
        iproute2 \
        iputils-ping \
        jq \
        less \
        lv \
        mysql-client \
        net-tools \
        openssh-client \
        s3cmd \
        strace \
        sysstat \
        tcpdump \
        telnet \
        traceroute \
        vim \
    && rm -rf /var/lib/apt/lists/* \
    && setcap -r /usr/bin/ping

COPY install-necocli /usr/local/bin

USER 10000
