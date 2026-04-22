# syntax=docker/dockerfile:1

FROM scratch AS argocd-image-updater-src
ARG ARGOCD_IMAGE_UPDATER_VERSION=1.1.1
ADD --checksum=sha256:dd30c3fc3730e965f8eeae8acdc363b3733c5994625004d5f582b8863abb4c9b \
    https://github.com/argoproj-labs/argocd-image-updater/archive/refs/tags/v${ARGOCD_IMAGE_UPDATER_VERSION}.tar.gz /src.tar.gz

FROM scratch AS awscli-download-amd64
ADD --checksum=sha256:5d08619f44becd0afcdd8d9a44b2a3481f4ee291d7bb5e06f767fabc7a9db22a \
    https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip /awscliv2.zip
ADD --checksum=sha256:cfe7c93250fa8bba7ae4ab1668ce3f456a0875a3c640b3012cdd2a36681b6225 \
    https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip.sig /awscliv2.sig

FROM scratch AS awscli-download-arm64
ADD --checksum=sha256:949f992854bd48ffd2e4aad020a91f1e977bff1f7225895b98e0d80d70b688d0 \
    https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip /awscliv2.zip
ADD --checksum=sha256:879121abcbe1fddc97271b7186cb13bb0b4a11a2aea38b60ecff2876b3112fea \
    https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip.sig /awscliv2.sig

FROM awscli-download-${TARGETARCH} AS awscli-download

FROM ghcr.io/cybozu/golang:1.25.9.2_noble@sha256:0aca7f1b3326079d444b33478612421cd6b7b087f50e0a2bd0335173095a1b22 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /work/argocd-image-updater

RUN --mount=type=bind,from=argocd-image-updater-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .

RUN mkdir -p dist && \
    OS=${TARGETOS} ARCH=${TARGETARCH} make build

FROM ghcr.io/cybozu/ubuntu:24.04.20260422@sha256:eb1812859faf05ef7ae1320383c6e09356a44d00521d2c6204daddcb507a4231
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
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o "awscliv2.zip" && \
    curl "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip.sig" -o "awscliv2.sig" && \
    # Verify
    gpg --import awscli-gpg-public-key.pub && \
    gpg --verify awscliv2.sig awscliv2.zip && \
    ## Install
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf awscliv2.zip ./aws

COPY --from=builder /work/argocd-image-updater/dist/argocd-image-updater /manager
COPY --from=builder /work/argocd-image-updater/hack/git-ask-pass.sh /usr/local/bin/git-ask-pass.sh
COPY --from=builder /work/argocd-image-updater/LICENSE /usr/local/argocd-image-updater/LICENSE

USER 10000:10000

ENTRYPOINT ["/usr/bin/tini", "--", "/manager"]
