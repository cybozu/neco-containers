FROM quay.io/cybozu/ubuntu:20.04

ARG KUBECTL_VERSION=1.22.9

RUN curl -L https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
    -o /usr/bin/kubectl && chmod a+x /usr/bin/kubectl

USER 10000:10000

ENTRYPOINT [ "/bin/bash" ]
