FROM ghcr.io/cybozu/golang:1.23-jammy AS builder

ENV ARGOCD_IMAGE_UPDATER_VERSION=0.15.2

WORKDIR /work

RUN git clone --depth 1 --branch v${ARGOCD_IMAGE_UPDATER_VERSION} https://github.com/argoproj-labs/argocd-image-updater.git

WORKDIR /work/argocd-image-updater
RUN mkdir -p dist && \
    make controller

FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN groupadd -g 10000 argocd && \
    useradd -r -u 10000 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        git \
        gpg \
        gpg-agent \
        openssh-client \
        tini \
        unzip && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o "awscliv2.zip" \
    && unzip awscliv2.zip \
    && ./aws/install \
    && rm -rf awscliv2.zip ./aws

COPY --from=builder /work/argocd-image-updater/dist/argocd-image-updater /usr/local/bin
COPY --from=builder /work/argocd-image-updater/hack/git-ask-pass.sh /usr/local/bin/git-ask-pass.sh
COPY --from=builder /work/argocd-image-updater/LICENSE /usr/local/argocd-image-updater/LICENSE

USER 10000:10000

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/argocd-image-updater"]
