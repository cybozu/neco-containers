# Argo CD container
# Refer to this Dockerfile: https://github.com/argoproj/argo-cd/blob/v0.12.1/Dockerfile

# Stage1: Initial Stage which pulls prepares build dependencies and CLI tooling we need for out final image
FROM quay.io/cybozu/golang:1.12-bionic AS builder
WORKDIR /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# Install dep
ENV DEP_VERSION=0.5.0
RUN curl -sSLf https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 -o /usr/local/bin/dep && \
    chmod +x /usr/local/bin/dep

# Install packr
ENV PACKR_VERSION=1.21.9
RUN curl -sSLf https://github.com/gobuffalo/packr/releases/download/v${PACKR_VERSION}/packr_${PACKR_VERSION}_linux_amd64.tar.gz | \
    tar zxf - -C /tmp/ && \
    mv /tmp/packr /usr/local/bin/packr

# Install kubectl
# NOTE: keep the version synced with https://storage.googleapis.com/kubernetes-release/release/stable.txt
ENV KUBECTL_VERSION=1.14.3
RUN curl -sSLf -o /usr/local/bin/kubectl -LO https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

# Install ksonnet
ENV KSONNET_VERSION=0.13.1
RUN curl -sSLf https://github.com/ksonnet/ksonnet/releases/download/v${KSONNET_VERSION}/ks_${KSONNET_VERSION}_linux_amd64.tar.gz | \
    tar zxf - -C /tmp/ && \
    mv /tmp/ks_${KSONNET_VERSION}_linux_amd64/ks /usr/local/bin/ks

# Install kustomize
ENV KUSTOMIZE_VERSION=2.0.3
RUN curl -sSLf -o /usr/local/bin/kustomize https://github.com/kubernetes-sigs/kustomize/releases/download/v${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64 && \
    chmod +x /usr/local/bin/kustomize


# Stage2; Argo CD Build Stage
FROM quay.io/cybozu/golang:1.12-bionic AS argocd-build
ENV ARGOCD_VERSION=1.1.0-rc1
ENV PACKAGE=github.com/argoproj/argo-cd

COPY --from=builder /usr/local/bin/dep /usr/local/bin/dep
COPY --from=builder /usr/local/bin/packr /usr/local/bin/packr

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# A dummy directory is created under $GOPATH/src/dummy so we are able to use dep
# to install all the packages of our dep lock file
RUN mkdir -p "${GOPATH}/src/dummy" \
    && curl -sSLf -o "${GOPATH}/src/dummy/Gopkg.toml https://raw.githubusercontent.com/argoproj/argo-cd/v${ARGOCD_VERSION}/Gopkg.toml" \
    && curl -sSLf -o "${GOPATH}/src/dummy/Gopkg.lock https://raw.githubusercontent.com/argoproj/argo-cd/v${ARGOCD_VERSION}/Gopkg.lock"

WORKDIR ${GOPATH}/src/dummy
RUN dep ensure -vendor-only && \
    mv vendor/* "${GOPATH}/src/" && \
    rmdir vendor

RUN curl -sSLf "https://${PACKAGE}/archive/v${ARGOCD_VERSION}.tar.gz" | \
    tar zxf - -C /work/ \
    && mkdir -p "${GOPATH}/src/github.com/argoproj/" \
    && mv "argo-cd-${ARGOCD_VERSION}" "${GOPATH}/src/${PACKAGE}"

# Perform the build
WORKDIR /go/src/github.com/argoproj/argo-cd
RUN make cli server controller repo-server argocd-util

# Final image
FROM quay.io/cybozu/ubuntu:18.04
RUN groupadd -g 10000 argocd && \
    useradd -r -u 10000 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd && \
    apt-get update && \
    apt-get install -y --no-install-recommends git && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /usr/local/bin/ks /usr/local/bin/ks
COPY --from=builder /usr/local/bin/kubectl /usr/local/argocd/bin/kubectl
COPY --from=builder /usr/local/bin/kustomize /usr/local/argocd/bin/kustomize

# workaround ksonnet issue https://github.com/ksonnet/ksonnet/issues/298
ENV USER=argocd

COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/dist/* /usr/local/argocd/bin/

COPY install-tools /usr/local/argocd/install-tools
ENV PATH=/usr/local/argocd/bin:"$PATH"

USER 10000:10000

WORKDIR /home/argocd
