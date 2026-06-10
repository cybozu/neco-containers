# syntax=docker/dockerfile:1

ARG ARGOCD_IMAGE_UPDATER_VERSION=1.2.1
# Argocd image updater requires the git commit hash in Makefile.
# https://github.com/argoproj-labs/argocd-image-updater/blob/1b622412d7ecc2392e86bd5d1d2fee755fc71ca1/Makefile#L18
ARG ARGOCD_IMAGE_UPDATER_GIT_COMMIT=1b622412d7ecc2392e86bd5d1d2fee755fc71ca1
ARG AWSCLI_VERSION=2.34.63

FROM scratch AS argocd-image-updater-src
ARG ARGOCD_IMAGE_UPDATER_VERSION
ADD --checksum=sha256:3fd0d9eca433fc01a77f519a56f925e5f2cb902415ea4aa68fd6b0e44eee0a9b \
    https://github.com/argoproj-labs/argocd-image-updater/archive/refs/tags/v${ARGOCD_IMAGE_UPDATER_VERSION}.tar.gz /src.tar.gz

FROM scratch AS awscli-download-amd64
ARG AWSCLI_VERSION
ADD --checksum=sha256:bc15b053ef5710be63c9d8c54089ec6847737457a936a10b120a0a4d08c7c771 \
    https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_VERSION}.zip /awscliv2.zip
ADD --checksum=sha256:6225bc72f040d9734d78ce47f20781cf3f14fc22db542a37a09c8c77b69cbc86 \
    https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_VERSION}.zip.sig /awscliv2.sig

FROM scratch AS awscli-download-arm64
ARG AWSCLI_VERSION
ADD --checksum=sha256:23ed2f830b3be601ad0d8f9c29c7f4a3695a555486de7ed24dff45df9e0860d1 \
    https://awscli.amazonaws.com/awscli-exe-linux-aarch64-${AWSCLI_VERSION}.zip /awscliv2.zip
ADD --checksum=sha256:9f6a7e6adf1da500522270735c63199deb7dcf83a3a5367eb197d01f505b4b5c \
    https://awscli.amazonaws.com/awscli-exe-linux-aarch64-${AWSCLI_VERSION}.zip.sig /awscliv2.sig

FROM awscli-download-${TARGETARCH} AS awscli-download

FROM ghcr.io/cybozu/golang:1.26.4.1_noble@sha256:add9d704d4b75df2c51328615be89b61a3e71e4833321aa02c3f325a30d3eb8f AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG ARGOCD_IMAGE_UPDATER_GIT_COMMIT

WORKDIR /work/argocd-image-updater

RUN --mount=type=bind,from=argocd-image-updater-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .

RUN mkdir -p dist && \
    OS=${TARGETOS} ARCH=${TARGETARCH} make build GIT_COMMIT=${ARGOCD_IMAGE_UPDATER_GIT_COMMIT}

FROM ghcr.io/cybozu/ubuntu:24.04.20260604@sha256:c7c519ae7504df61e01495635ebda9c8698f2d221cd3fcf6e9500ad97b9d2421
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN groupadd -g 10000 argocd && \
    useradd -r -u 10000 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        git \
        gpg \
        gpg-agent \
        openssh-client \
        tini \
        unzip && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# awscli
WORKDIR /work
COPY awscli-gpg-public-key.pub .
RUN --mount=type=bind,from=awscli-download,source=/awscliv2.zip,target=/work/awscliv2.zip \
    --mount=type=bind,from=awscli-download,source=/awscliv2.sig,target=/work/awscliv2.sig \
    # Verify
    gpg --import awscli-gpg-public-key.pub && \
    gpg --verify awscliv2.sig awscliv2.zip && \
    ## Install
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf ./aws

COPY --from=builder /work/argocd-image-updater/dist/argocd-image-updater /manager
COPY --from=builder /work/argocd-image-updater/hack/git-ask-pass.sh /usr/local/bin/git-ask-pass.sh
COPY --from=builder /work/argocd-image-updater/LICENSE /usr/local/argocd-image-updater/LICENSE

USER 10000:10000

ENTRYPOINT ["/usr/bin/tini", "--", "/manager"]
