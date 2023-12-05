FROM quay.io/cybozu/golang:1.21-jammy AS builder

ENV ARGOCD_IMAGE_UPDATER_VERSION=0.12.2

WORKDIR /work

RUN git clone --depth 1 --branch v${ARGOCD_IMAGE_UPDATER_VERSION} https://github.com/argoproj-labs/argocd-image-updater.git

WORKDIR /work/argocd-image-updater
RUN mkdir -p dist && \
    make controller

FROM quay.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN groupadd -g 10000 argocd && \
    useradd -r -u 10000 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd && \
    apt-get update && \
    apt-get install -y --no-install-recommends git openssh-client python3 python3-pip && \
    pip3 install --upgrade pip && \
    pip3 install awscli && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /work/argocd-image-updater/dist/argocd-image-updater /usr/local/bin
COPY --from=builder /work/argocd-image-updater/hack/git-ask-pass.sh /usr/local/bin/git-ask-pass.sh
COPY --from=builder /work/argocd-image-updater/LICENSE /usr/local/argocd-image-updater/LICENSE

USER argocd

ENTRYPOINT ["/usr/local/bin/argocd-image-updater"]
