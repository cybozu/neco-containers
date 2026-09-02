# syntax=docker/dockerfile:1
# Argo CD container

ARG ARGOCD_VERSION=3.5.2
# Workaround: upstream uses Kustomize v5.8.1 which introduces breaking changes, so we use v5.7.1 which keeps backward compatibilities.
#             See. https://github.com/kubernetes-sigs/kustomize/issues/6058
ARG KUSTOMIZE_VERSION=5.7.1
ARG HELM_VERSION=4.2.1
ARG GIT_LFS_VERSION=3.7.1
ARG NODE_VERSION=24.20.0
ARG PNPM_VERSION=10.34.5

FROM scratch AS kustomize-download
ARG KUSTOMIZE_VERSION
ADD --checksum=sha256:ea375e7372f9aa029129d4b2d16c66b7750b7f1213c4f66f910d981c895818d8 \
    https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz /kustomize.tar.gz

FROM scratch AS helm-download
ARG HELM_VERSION
ADD --checksum=sha256:479dca836e5b45e8bd222400c5591b0e3a647378f03ff96597180db97c17fdae \
    https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz /helm.tar.gz

FROM scratch AS git-lfs-download
ARG GIT_LFS_VERSION
ADD --checksum=sha256:1c0b6ee5200ca708c5cebebb18fdeb0e1c98f1af5c1a9cba205a4c0ab5a5ec08 \
    https://github.com/git-lfs/git-lfs/releases/download/v${GIT_LFS_VERSION}/git-lfs-linux-amd64-v${GIT_LFS_VERSION}.tar.gz /git-lfs.tar.gz

FROM scratch AS node-download
ARG NODE_VERSION
ADD --checksum=sha256:855d581f8a4eb1a8117e3426de25fe02770592febcfb31369aee1ffbfee9e8ec \
    https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz /node.tar.gz

FROM scratch AS argocd-src
ARG ARGOCD_VERSION
ADD --checksum=sha256:8236a02fefd9cc349b6bf2ce554713c47a58e3f4ce498ee0208504c8873ef2a8 \
    https://github.com/argoproj/argo-cd/archive/refs/tags/v${ARGOCD_VERSION}.tar.gz /src.tar.gz

FROM scratch AS pnpm-download
ARG PNPM_VERSION
ADD --checksum=sha256:69a535476f407fb0f64427913c2b592ca2ed6bf93f92e2cb82a32349e31f36eb \
    https://github.com/pnpm/pnpm/releases/download/v${PNPM_VERSION}/pnpm-linux-x64 /pnpm

FROM ghcr.io/cybozu/ubuntu:24.04.20260902@sha256:182a16198fafecbce2f813a07de269d28efb885aa4843e611ce44c25ce46cc2b AS cybozu-ubuntu
    
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
FROM ghcr.io/cybozu/golang:1.26.8.1_noble@sha256:fcdd7870a523950129a6afc09e03c035796446ac7c1b2f84994e4a7920042879 AS argocd-build
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
    pnpm config set -g minimumReleaseAge 1440

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
