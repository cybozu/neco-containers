# Go container

FROM --platform=$TARGETPLATFORM ghcr.io/cybozu/ubuntu-dev:20.04
LABEL org.opencontainers.image.source "https://github.com/cybozu/neco-containers"

ARG TARGETARCH
ARG GO_VERSION=1.20.11

ENV GOARCH=${TARGETARCH}
ENV GOPATH=/go
ENV PATH=/go/bin:/usr/local/go/bin:"$PATH"

RUN rm -rf /usr/local/go \
    && curl -sfL https://dl.google.com/go/go${GO_VERSION}.linux-${GOARCH}.tar.gz \
    | tar -x -z -C /usr/local -f - \
    && mkdir -p /go/src \
    && GOBIN=/usr/local/bin go install golang.org/x/tools/cmd/goimports@latest \
    && GOBIN=/usr/local/bin go install golang.org/x/lint/golint@latest \
    && GOBIN=/usr/local/bin go install honnef.co/go/tools/cmd/staticcheck@latest \
    && GOBIN=/usr/local/bin go install github.com/gordonklaus/ineffassign@latest \
    && GOBIN=/usr/local/bin go install github.com/tcnksm/ghr@latest \
    && GOBIN=/usr/local/bin go install github.com/cybozu-go/golang-custom-analyzer/cmd/...@latest \
    && rm -rf /go \
    && mkdir -p /go/src

WORKDIR /work

CMD ["/bin/bash"]
