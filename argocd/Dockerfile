# Argo CD container

# Stage1: Initial Stage which pulls prepares build dependencies and CLI tooling we need for out final image
FROM quay.io/cybozu/ubuntu:20.04 AS builder
WORKDIR /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install kustomize
ENV KUSTOMIZE_VERSION=4.4.1
RUN curl -sSLf https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz | \
    tar zxf - -C /usr/local/bin

# Install helm
ENV HELM_VERSION=3.8.0
RUN curl -sSLf https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz | \
    tar zxf - -C /usr/local/bin --strip-components 1

# Stage2; Argo CD Build Stage
FROM quay.io/cybozu/golang:1.17-focal AS argocd-build
ENV ARGOCD_VERSION=2.3.9
ENV PACKAGE=github.com/argoproj/argo-cd

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install Node.js 12.x and yarn
# https://github.com/nodesource/distributions/blob/master/README.md#debinstall
RUN curl -sSLf https://deb.nodesource.com/setup_12.x | bash - \
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
FROM quay.io/cybozu/ubuntu:20.04
RUN groupadd -g 10000 argocd && \
    useradd -r -u 10000 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd && \
    apt-get update && \
    apt-get install -y --no-install-recommends git gpg gpg-agent openssh-client && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /usr/local/bin/kustomize /usr/local/bin/kustomize
COPY --from=builder /usr/local/bin/helm /usr/local/bin/helm

# workaround ksonnet issue https://github.com/ksonnet/ksonnet/issues/298
ENV USER=argocd

COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/dist/* /usr/local/bin/
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/gpg-wrapper.sh /usr/local/bin/gpg-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/git-verify-wrapper.sh /usr/local/bin/git-verify-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/LICENSE /usr/local/argocd/LICENSE

RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-server
RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-repo-server
RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-cmp-server
RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-application-controller
RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-dex
RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-notifications

USER 10000:10000

WORKDIR /home/argocd
