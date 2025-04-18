# Argo CD container

# Stage1: Initial Stage which pulls prepares build dependencies and CLI tooling we need for out final image
FROM ghcr.io/cybozu/ubuntu:22.04 AS builder
WORKDIR /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install kustomize
ENV KUSTOMIZE_VERSION=5.4.3
RUN curl -sSLf https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz | \
    tar zxf - -C /usr/local/bin

# Install helm
ENV HELM_VERSION=3.15.4
RUN curl -sSLf https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz | \
    tar zxf - -C /usr/local/bin --strip-components 1

# Install tini
ARG TINI_VERSION=0.19.0
RUN curl -fsSL -o /usr/local/bin/tini-static https://github.com/krallin/tini/releases/download/v${TINI_VERSION}/tini-static && \
    chmod +x /usr/local/bin/tini-static

# Stage2; Argo CD Build Stage
FROM ghcr.io/cybozu/golang:1.23-jammy AS argocd-build
ENV ARGOCD_VERSION=2.13.2
ENV PACKAGE=github.com/argoproj/argo-cd

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install Node.js 20.x and yarn
# https://github.com/nodesource/distributions/blob/master/README.md#debinstall
RUN curl -sSLf https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global yarn

WORKDIR /work
RUN curl -sSLf "https://${PACKAGE}/archive/v${ARGOCD_VERSION}.tar.gz" | \
    tar zxf - -C /work/ \
    && mkdir -p "${GOPATH}/src/github.com/argoproj/" \
    && mv "argo-cd-${ARGOCD_VERSION}" "${GOPATH}/src/${PACKAGE}"

# UI stage
WORKDIR /go/src/github.com/argoproj/argo-cd/ui
RUN yarn install --frozen-lockfile

ENV ARGO_VERSION=$ARGOCD_VERSION
RUN NODE_ENV='production' NODE_ONLINE_ENV='online' yarn build

# Perform the build
WORKDIR /go/src/github.com/argoproj/argo-cd
RUN make argocd-all


# Final image
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN groupadd -g 10000 argocd && \
    useradd -r -u 10000 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd && \
    apt-get update && \
    apt-get install -y --no-install-recommends git git-lfs gpg gpg-agent openssh-client connect-proxy && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /usr/local/bin/tini-static /usr/local/bin/tini
COPY --from=builder /usr/local/bin/kustomize /usr/local/bin/kustomize
COPY --from=builder /usr/local/bin/helm /usr/local/bin/helm

# workaround ksonnet issue https://github.com/ksonnet/ksonnet/issues/298
ENV USER=argocd

COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/dist/* /usr/local/bin/
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/gpg-wrapper.sh /usr/local/bin/gpg-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/git-verify-wrapper.sh /usr/local/bin/git-verify-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/LICENSE /usr/local/argocd/LICENSE

RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-repo-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-cmp-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-application-controller && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-dex && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-notifications

USER 10000:10000

WORKDIR /home/argocd

ENTRYPOINT ["tini", "--"]
