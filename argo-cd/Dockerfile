# Argo CD container
# Refer to this Dockerfile: https://github.com/argoproj/argo-cd/blob/v0.11.1/Dockerfile

# Stage1: Initial Stage which pulls prepares build dependencies and CLI tooling we need for out final image
FROM quay.io/cybozu/golang:1.11-bionic AS builder
WORKDIR /tmp

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
# Keep version at 1.12.X until https://github.com/argoproj/argo-cd/issues/1012 is resolved
ENV KUBECTL_VERSION=1.12.4
RUN curl -sSLf -o /usr/local/bin/kubectl -LO https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

# Install kustomize
ENV KUSTOMIZE_VERSION=1.0.11
RUN curl -sSLf -o /usr/local/bin/kustomize https://github.com/kubernetes-sigs/kustomize/releases/download/v${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64 && \
    chmod +x /usr/local/bin/kustomize


# Stage2; Argo CD Build Stage
FROM quay.io/cybozu/golang:1.11-bionic AS argocd-build
ENV ARGOCD_VERSION=0.11.1
ENV PACKAGE=github.com/argoproj/argo-cd

RUN curl -sSLf https://${PACKAGE}/archive/v${ARGOCD_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mkdir -p ${GOPATH}/src/github.com/argoproj/ \
    && mv argo-cd-${ARGOCD_VERSION} ${GOPATH}/src/${PACKAGE}

COPY --from=builder /usr/local/bin/dep /usr/local/bin/dep
COPY --from=builder /usr/local/bin/packr /usr/local/bin/packr

# A dummy directory is created under $GOPATH/src/dummy so we are able to use dep
# to install all the packages of our dep lock file
COPY Gopkg.toml ${GOPATH}/src/dummy/Gopkg.toml
COPY Gopkg.lock ${GOPATH}/src/dummy/Gopkg.lock

RUN cd ${GOPATH}/src/dummy && \
    dep ensure -vendor-only && \
    mv vendor/* ${GOPATH}/src/ && \
    rmdir vendor

# Perform the build
WORKDIR /go/src/github.com/argoproj/argo-cd
COPY Makefile .
RUN make cli server controller repo-server argocd-util


# Final image
FROM quay.io/cybozu/ubuntu:18.04
RUN groupadd -g 999 argocd && \
    useradd -r -u 999 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:argocd /home/argocd

COPY --from=builder /usr/local/bin/kubectl /usr/local/argocd/bin/kubectl
COPY --from=builder /usr/local/bin/kustomize /usr/local/argocd/bin/kustomize

# workaround ksonnet issue https://github.com/ksonnet/ksonnet/issues/298
ENV USER=argocd

COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/dist/* /usr/local/argocd/bin/

COPY install-tools /usr/local/argocd/bin/install-tools
ENV PATH=/usr/local/argocd/bin:"$PATH"

USER argocd

WORKDIR /home/argocd
