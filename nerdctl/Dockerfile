FROM quay.io/cybozu/ubuntu:20.04 AS build

ARG NERDCTL_VERSION=0.20.0
RUN curl -sSL --output /tmp/nerdctl.tgz https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-${TARGETARCH:-amd64}.tar.gz  \
	&& tar zxvf /tmp/nerdctl.tgz -C /usr/local/bin/ \
	&&     rm /tmp/nerdctl.tgz 
ENTRYPOINT ["/usr/local/bin/nerdctl", "ipfs", "registry", "serve"]
