# syntax=docker/dockerfile:1
# Argo CD container

ARG ARGOCD_VERSION=3.5.4
# Workaround: upstream uses Kustomize v5.8.1 which introduces breaking changes, so we use v5.7.1 which keeps backward compatibilities.
#             See. https://github.com/kubernetes-sigs/kustomize/issues/6058
ARG KUSTOMIZE_VERSION=5.7.1
# Workaround: upstream uses Helm v4.2.1 which introduces breaking changes in conjunction with kustomize before v5.8.0, so we use v3.22.0 which keeps backward compatibilities.
#             See. https://github.com/kubernetes-sigs/kustomize/issues/6013
ARG HELM_VERSION=3.22.0
ARG GIT_LFS_VERSION=3.7.1
ARG NODE_VERSION=24.21.0
ARG PNPM_VERSION=10.34.6

FROM scratch AS kustomize-download
ARG KUSTOMIZE_VERSION
ADD --checksum=sha256:ea375e7372f9aa029129d4b2d16c66b7750b7f1213c4f66f910d981c895818d8 \
    https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz /kustomize.tar.gz

FROM scratch AS helm-download
ARG HELM_VERSION
ADD --checksum=sha256:1e4ab49e429626cf6c6958d914248b78c9730803c2751b87627e171dc800e7bb \
    https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz /helm.tar.gz

FROM scratch AS git-lfs-download
ARG GIT_LFS_VERSION
ADD --checksum=sha256:1c0b6ee5200ca708c5cebebb18fdeb0e1c98f1af5c1a9cba205a4c0ab5a5ec08 \
    https://github.com/git-lfs/git-lfs/releases/download/v${GIT_LFS_VERSION}/git-lfs-linux-amd64-v${GIT_LFS_VERSION}.tar.gz /git-lfs.tar.gz

FROM scratch AS node-download
ARG NODE_VERSION
ADD --checksum=sha256:6e1db87ef58b8819e5d5402eff1536491b18edd8eb7bee5ef7897876e88dc5ff \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz /node.tar.gz

FROM scratch AS argocd-src
ARG ARGOCD_VERSION
ADD --checksum=sha256:3c346ebbc871ac5d22d1fea01b83bf26905594b85f6d93ebe6b1cf64def10826 \
    https://github.com/argoproj/argo-cd/archive/refs/tags/v${ARGOCD_VERSION}.tar.gz /src.tar.gz

FROM scratch AS pnpm-download
ARG PNPM_VERSION
ADD --checksum=sha256:1bdfa04c3acbbf019befa4cf51f73f1de2729f4a513f737d1295905277f3b9cf \
    https://github.com/pnpm/pnpm/releases/download/v${PNPM_VERSION}/pnpm-linux-x64 /pnpm

FROM ghcr.io/cybozu/ubuntu:24.04.20261005@sha256:848752a739357921bbbe0780e9f86cb91ffcbeed490882ac550e30e425063be1 AS cybozu-ubuntu

# Stage1: Initial Stage which prepares build dependencies and CLI tooling we need for out final image
FROM cybozu-ubuntu AS builder
WORKDIR /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install kustomize
RUN --mount=type=bind,from=kustomize-download,source=/kustomize.tar.gz,target=/tmp/kustomize.tar.gz \
    tar xf /tmp/kustomize.tar.gz -C /usr/local/bin

# Install helm
RUN --mount=type=bind,from=helm-download,source=/helm.tar.gz,target=/tmp/helm.tar.gz \
    tar xf /tmp/helm.tar.gz -C /usr/local/bin --strip-components 1

# Install git-lfs
RUN --mount=type=bind,from=git-lfs-download,source=/git-lfs.tar.gz,target=/tmp/git-lfs.tar.gz \
    tar xf /tmp/git-lfs.tar.gz -C /usr/local/bin --strip-components 1

# Stage2; Argo CD Build Stage
FROM ghcr.io/cybozu/golang:1.27.1.1_noble@sha256:e38fe3b72f61d034394ee2c2592d41fa753226718bbc111bb6bb9a21601f0859 AS argocd-build
ARG ARGOCD_VERSION
ENV ARGOCD_VERSION=${ARGOCD_VERSION}
ENV PACKAGE=github.com/argoproj/argo-cd

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Use stable version instead of unstable version upstream uses
ARG NODE_VERSION
RUN --mount=type=bind,from=node-download,source=/node.tar.gz,target=/tmp/node.tar.gz \
    tar xf /tmp/node.tar.gz -C /opt
ENV PATH="/opt/node-v${NODE_VERSION}-linux-x64/bin:${PATH}"

RUN --mount=type=bind,from=pnpm-download,source=/pnpm,target=/tmp/pnpm \
    install -m 0755 /tmp/pnpm /usr/local/bin/pnpm && \
    # 3 days, per the company-wide npm / pnpm guideline. Only used if pnpm has to
    # resolve versions; the frozen-lockfile install below does not resolve anything.
    pnpm config set -g minimumReleaseAge 4320

RUN apt-get update && apt-get install -y --no-install-recommends gpg && \
    # Unlike upstream, do not remove /usr/share/doc/* to preserve license files
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /go/src/github.com/argoproj/argo-cd
RUN --mount=type=bind,from=argocd-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xf /tmp/src.tar.gz --strip-components=1 -C .

# UI stage
WORKDIR /go/src/github.com/argoproj/argo-cd/ui
RUN pnpm install --frozen-lockfile

ENV ARGO_VERSION=${ARGOCD_VERSION}
RUN NODE_ENV='production' NODE_ONLINE_ENV='online' NODE_OPTIONS=--max_old_space_size=8192 pnpm build

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
        gpg \
        gpg-agent \
        openssh-client \
        tini && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /usr/local/bin/kustomize /usr/local/bin/kustomize
COPY --from=builder /usr/local/bin/helm /usr/local/bin/helm
COPY --from=builder /usr/local/bin/git-lfs /usr/local/bin/git-lfs
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/dist/* /usr/local/bin/
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/gpg-wrapper.sh /usr/local/bin/gpg-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/git-verify-wrapper.sh /usr/local/bin/git-verify-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/LICENSE /usr/local/argocd/LICENSE

RUN git lfs install --system

WORKDIR /app/config
RUN mkdir -p tls && \
    mkdir -p gpg/source && \
    mkdir -p gpg/keys && \
    chown argocd gpg/keys && \
    chmod 0700 gpg/keys

# Disable gRPC service config lookups via DNS TXT records to prevent excessive
# DNS queries for _grpc_config.<hostname> which can cause timeouts in dual-stack
# environments. This can be overridden via argocd-cmd-params-cm ConfigMap.
# See https://github.com/argoproj/argo-cd/issues/24991
ENV GRPC_ENABLE_TXT_SERVICE_CONFIG=false

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
