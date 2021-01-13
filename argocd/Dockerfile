# Argo CD container
# Refer to this Dockerfile: https://github.com/argoproj/argo-cd/blob/v0.12.1/Dockerfile

# Stage1: Initial Stage which pulls prepares build dependencies and CLI tooling we need for out final image
FROM quay.io/cybozu/golang:1.15-focal AS builder
WORKDIR /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# Install packr
ENV PACKR_VERSION=1.21.9
RUN curl -sSLf https://github.com/gobuffalo/packr/releases/download/v${PACKR_VERSION}/packr_${PACKR_VERSION}_linux_amd64.tar.gz | \
    tar zxf - -C /tmp/ && \
    mv /tmp/packr /usr/local/bin/packr

# Install kubectl
# NOTE: keep the version synced with https://storage.googleapis.com/kubernetes-release/release/stable.txt
ENV KUBECTL_VERSION=1.19.7
RUN curl -sSLf -o /usr/local/bin/kubectl -LO https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

# Install kustomize
ENV KUSTOMIZE_VERSION=3.7.0
RUN curl -sSLf https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz | \
    tar zxf - -C /usr/local/bin

# Install helm
ENV HELM_VERSION=3.4.2
RUN curl -sSLf https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz | \
    tar zxf - -C /usr/local/bin --strip-components 1

# Stage2; Argo CD Build Stage
FROM quay.io/cybozu/golang:1.15-focal AS argocd-build
ENV ARGOCD_VERSION=1.7.7
ENV PACKAGE=github.com/argoproj/argo-cd

COPY --from=builder /usr/local/bin/packr /usr/local/bin/packr

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /work
RUN curl -sSLf "https://${PACKAGE}/archive/v${ARGOCD_VERSION}.tar.gz" | \
    tar zxf - -C /work/ \
    && mkdir -p "${GOPATH}/src/github.com/argoproj/" \
    && mv "argo-cd-${ARGOCD_VERSION}" "${GOPATH}/src/${PACKAGE}"

# Perform the build
WORKDIR /go/src/github.com/argoproj/argo-cd
RUN make cli-local server controller repo-server argocd-util

# Install yarn for building Web UI
RUN apt-get update && apt-get install -y --no-install-recommends gnupg \
    && curl -sLf https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
    && echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list \
    && apt-get update && apt-get install -y --no-install-recommends yarn nodejs \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/argoproj/argo-cd/ui
RUN yarn install

ENV ARGO_VERSION=$ARGOCD_VERSION
RUN NODE_ENV='production' yarn build


# Final image
FROM quay.io/cybozu/ubuntu:20.04
RUN groupadd -g 10000 argocd && \
    useradd -r -u 10000 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd && \
    apt-get update && \
    apt-get install -y --no-install-recommends git gpg gpg-agent && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /usr/local/bin/kubectl /usr/local/argocd/bin/kubectl
COPY --from=builder /usr/local/bin/kustomize /usr/local/argocd/bin/kustomize
COPY --from=builder /usr/local/bin/helm /usr/local/argocd/bin/helm

# workaround ksonnet issue https://github.com/ksonnet/ksonnet/issues/298
ENV USER=argocd

COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/dist/* /usr/local/argocd/bin/
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/ui/dist/app /shared/argocd/app
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/git-ask-pass.sh /usr/local/argocd/bin/git-ask-pass.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/gpg-wrapper.sh /usr/local/bin/gpg-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/hack/git-verify-wrapper.sh /usr/local/bin/git-verify-wrapper.sh
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/LICENSE /usr/local/argocd/LICENSE

COPY install-tools /usr/local/argocd/install-tools
ENV PATH=/usr/local/argocd/bin:"$PATH"

USER 10000:10000

WORKDIR /home/argocd
