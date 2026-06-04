# syntax=docker/dockerfile:1
# Argo CD container

ARG ARGOCD_VERSION=3.4.3
# Workaround: upstream uses Kustomize v5.8.1 which introduces breaking changes, so we use v5.7.1 which keeps backward compatibilities.
#             See. https://github.com/kubernetes-sigs/kustomize/issues/6058
ARG KUSTOMIZE_VERSION=5.7.1
ARG HELM_VERSION=3.19.4
ARG NODE_VERSION=24.16.0

FROM scratch AS kustomize-download
ARG KUSTOMIZE_VERSION
ADD --checksum=sha256:ea375e7372f9aa029129d4b2d16c66b7750b7f1213c4f66f910d981c895818d8 \
    https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz /kustomize.tar.gz

FROM scratch AS helm-download
ARG HELM_VERSION
ADD --checksum=sha256:759c656fbd9c11e6a47784ecbeac6ad1eb16a9e76d202e51163ab78504848862 \
    https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz /helm.tar.gz

FROM scratch AS node-download
ARG NODE_VERSION
ADD --checksum=sha256:2faf6a387e9b62b888e21c54f01249fb27537ffecf1842f29f4c919d0a59a0ff \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz /node.tar.gz

FROM scratch AS argocd-src
ARG ARGOCD_VERSION
ADD --checksum=sha256:4c14774e6107ab128546df8d891259ccd15ce738611b4a797e61d041ab2a2f29 \
    https://github.com/argoproj/argo-cd/archive/refs/tags/v${ARGOCD_VERSION}.tar.gz /src.tar.gz

FROM ghcr.io/cybozu/ubuntu:24.04.20260604@sha256:c7c519ae7504df61e01495635ebda9c8698f2d221cd3fcf6e9500ad97b9d2421 AS cybozu-ubuntu
    
# Stage1: Initial Stage which pulls prepares build dependencies and CLI tooling we need for out final image
FROM cybozu-ubuntu AS builder
WORKDIR /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install kustomize
RUN --mount=type=bind,from=kustomize-download,source=/kustomize.tar.gz,target=/tmp/kustomize.tar.gz \
    tar xvf /tmp/kustomize.tar.gz -C /usr/local/bin

# Install helm
RUN --mount=type=bind,from=helm-download,source=/helm.tar.gz,target=/tmp/helm.tar.gz \
    tar xvf /tmp/helm.tar.gz -C /usr/local/bin --strip-components 1

# Stage2; Argo CD Build Stage
FROM ghcr.io/cybozu/golang:1.26.4.1_noble@sha256:add9d704d4b75df2c51328615be89b61a3e71e4833321aa02c3f325a30d3eb8f AS argocd-build
ARG ARGOCD_VERSION
ENV ARGOCD_VERSION=${ARGOCD_VERSION}
ENV PACKAGE=github.com/argoproj/argo-cd

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Use stable version instead of unstable version upstream uses
ARG NODE_VERSION
RUN --mount=type=bind,from=node-download,source=/node.tar.gz,target=/tmp/node.tar.gz \
    tar xvf /tmp/node.tar.gz -C /opt
ENV PATH="/opt/node-v${NODE_VERSION}-linux-x64/bin:${PATH}"

RUN apt-get update && apt-get install -y --no-install-recommends gpg
# ref. https://classic.yarnpkg.com/lang/en/docs/install/#debian-stable
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | gpg --dearmor | tee /etc/apt/keyrings/yarn-archive-keyring.gpg > /dev/null && \
    echo "deb [signed-by=/etc/apt/keyrings/yarn-archive-keyring.gpg] https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get update && apt-get install -y --no-install-recommends yarn && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /go/src/github.com/argoproj/argo-cd
RUN --mount=type=bind,from=argocd-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar zxf /tmp/src.tar.gz --strip-components=1 -C .

# UI stage
WORKDIR /go/src/github.com/argoproj/argo-cd/ui
RUN yarn install --frozen-lockfile

ENV ARGO_VERSION=${ARGOCD_VERSION}
RUN NODE_ENV='production' NODE_ONLINE_ENV='online' NODE_OPTIONS=--max_old_space_size=8192 yarn build

# Perform the build
WORKDIR /go/src/github.com/argoproj/argo-cd
RUN make argocd-all

# Final image
FROM cybozu-ubuntu
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN groupadd -g 10000 argocd && \
    useradd -r -u 10000 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        connect-proxy \
        git \
        git-lfs \
        gpg \
        gpg-agent \
        openssh-client \
        tini && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /usr/local/bin/kustomize /usr/local/bin/kustomize
COPY --from=builder /usr/local/bin/helm /usr/local/bin/helm
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/dist/* /usr/local/bin/
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/gpg-wrapper.sh /usr/local/bin/gpg-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/git-verify-wrapper.sh /usr/local/bin/git-verify-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/LICENSE /usr/local/argocd/LICENSE

WORKDIR /app/config
RUN mkdir -p tls && \
    mkdir -p gpg/source && \
    mkdir -p gpg/keys && \
    chown argocd gpg/keys && \
    chmod 0700 gpg/keys

RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-repo-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-cmp-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-application-controller && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-dex && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-notifications && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-applicationset-controller && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-k8s-auth && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-commit-server

USER 10000:10000

WORKDIR /home/argocd

ENTRYPOINT ["tini", "--"]
