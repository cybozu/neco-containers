# syntax=docker/dockerfile:1

ARG ARGOCD_IMAGE_UPDATER_VERSION=1.3.0
# Argocd image updater requires the git commit hash in Makefile.
# https://github.com/argoproj-labs/argocd-image-updater/blob/1b622412d7ecc2392e86bd5d1d2fee755fc71ca1/Makefile#L18
ARG ARGOCD_IMAGE_UPDATER_GIT_COMMIT=098b544f1c7d6d1368e89c3f8d92415ede353916
ARG AWSCLI_VERSION=2.36.40

FROM scratch AS argocd-image-updater-src
ARG ARGOCD_IMAGE_UPDATER_VERSION
ADD --checksum=sha256:76634c40ba71619bcdf0441228e5c52303f03563ffb6650b32af683eb1ee3238 \
    https://github.com/argoproj-labs/argocd-image-updater/archive/refs/tags/v${ARGOCD_IMAGE_UPDATER_VERSION}.tar.gz /src.tar.gz

FROM scratch AS awscli-download-amd64
ARG AWSCLI_VERSION
ADD --checksum=sha256:a904e314340ad4c4b62beb50e9259d4becb2acf3dffc86c96c553c834d7fdf7a \
    https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_VERSION}.zip /awscliv2.zip
ADD --checksum=sha256:bb1cc030957e59c5681afad34ece0626536ac76881b703194ec0921d5b083204 \
    https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_VERSION}.zip.sig /awscliv2.sig

FROM scratch AS awscli-download-arm64
ARG AWSCLI_VERSION
ADD --checksum=sha256:f26e08d0d64d9f35f304a9df94d63cbd11ffe8881c8b4bf60be934cf9e4d13a9 \
    https://awscli.amazonaws.com/awscli-exe-linux-aarch64-${AWSCLI_VERSION}.zip /awscliv2.zip
ADD --checksum=sha256:a0c326843a091cb907849f18aa4c720ef5a59a9eaa541c20a4c1f61c9814619d \
    https://awscli.amazonaws.com/awscli-exe-linux-aarch64-${AWSCLI_VERSION}.zip.sig /awscliv2.sig

FROM awscli-download-${TARGETARCH} AS awscli-download

FROM ghcr.io/cybozu/golang:1.26.8.2_noble@sha256:e7d0eca1aa1de90597d2c7deeeaa7c93bc7605c51af380ccfc6776c892252a37 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG ARGOCD_IMAGE_UPDATER_GIT_COMMIT

WORKDIR /work/argocd-image-updater

RUN --mount=type=bind,from=argocd-image-updater-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .

RUN mkdir -p dist && \
    OS=${TARGETOS} ARCH=${TARGETARCH} make build GIT_COMMIT=${ARGOCD_IMAGE_UPDATER_GIT_COMMIT}

FROM ghcr.io/cybozu/ubuntu:24.04.20260902@sha256:182a16198fafecbce2f813a07de269d28efb885aa4843e611ce44c25ce46cc2b
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
