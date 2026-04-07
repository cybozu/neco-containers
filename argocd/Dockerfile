# Argo CD container

# Stage1: Initial Stage which pulls prepares build dependencies and CLI tooling we need for out final image
FROM ghcr.io/cybozu/ubuntu:24.04 AS builder
WORKDIR /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install kustomize
ARG KUSTOMIZE_VERSION=5.8.1
ARG KUSTOMIZE_SHA=029a7f0f4e1932c52a0476cf02a0fd855c0bb85694b82c338fc648dcb53a819d
RUN curl -fsSL -O https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz && \
    echo "${KUSTOMIZE_SHA}  kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz" | sha256sum -c - && \
    tar xvf kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz -C /usr/local/bin && \
    rm kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz

# Install helm
ARG HELM_VERSION=3.19.4
ARG HELM_SHA=759c656fbd9c11e6a47784ecbeac6ad1eb16a9e76d202e51163ab78504848862
RUN curl -fsSL -O https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz && \
    echo "${HELM_SHA}  helm-v${HELM_VERSION}-linux-amd64.tar.gz" | sha256sum -c - && \
    tar xvf helm-v${HELM_VERSION}-linux-amd64.tar.gz -C /usr/local/bin --strip-components 1 && \
    rm helm-v${HELM_VERSION}-linux-amd64.tar.gz

# Stage2; Argo CD Build Stage
FROM ghcr.io/cybozu/golang:1.24-noble AS argocd-build
ENV ARGOCD_VERSION=3.3.6
ENV PACKAGE=github.com/argoproj/argo-cd

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Use stable version instead of unstable version upstream uses
ARG NODE_VERSION=v24.14.1
ARG NODE_SHA=ace9fa104992ed0829642629c46ca7bd7fd6e76278cb96c958c4b387d29658ea
RUN curl -fsSL -O https://nodejs.org/dist/${NODE_VERSION}/node-${NODE_VERSION}-linux-x64.tar.gz && \
    echo "${NODE_SHA}  node-${NODE_VERSION}-linux-x64.tar.gz" | sha256sum -c - && \
    tar xvf node-${NODE_VERSION}-linux-x64.tar.gz -C /opt && \
    rm node-${NODE_VERSION}-linux-x64.tar.gz
ENV PATH="/opt/node-${NODE_VERSION}-linux-x64/bin:${PATH}"

RUN apt-get update && apt-get install -y --no-install-recommends gpg
# ref. https://classic.yarnpkg.com/lang/en/docs/install/#debian-stable
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | gpg --dearmor | tee /etc/apt/keyrings/yarn-archive-keyring.gpg > /dev/null && \
    echo "deb [signed-by=/etc/apt/keyrings/yarn-archive-keyring.gpg] https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get update && apt-get install -y --no-install-recommends yarn && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

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
