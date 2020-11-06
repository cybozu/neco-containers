# Go container

FROM quay.io/cybozu/ubuntu-dev:20.04

ARG GO_VERSION=1.15.4
WORKDIR /work

ENV GOPATH=/go
ENV PATH=/go/bin:/usr/local/go/bin:"$PATH"

RUN curl -s -f -O https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -x -z -C /usr/local -f go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz \
    && mkdir -p /go/src \
    && GOBIN=/usr/local/bin go get golang.org/x/tools/cmd/goimports \
    && GOBIN=/usr/local/bin go get golang.org/x/lint/golint \
    && GOBIN=/usr/local/bin go get honnef.co/go/tools/cmd/staticcheck \
    && GOBIN=/usr/local/bin go get github.com/gordonklaus/ineffassign \
    && GOBIN=/usr/local/bin go get github.com/tcnksm/ghr \
    && GOBIN=/usr/local/bin go get github.com/gostaticanalysis/nilerr/cmd/nilerr \
    && rm -rf /go/src \
    && mkdir -p /go/src

COPY ./analyzer /work/analyzer
WORKDIR /work/analyzer
RUN GOBIN=/usr/local/bin go install ./cmd/...

WORKDIR /work
RUN rm -r /work/analyzer

CMD ["/bin/bash"]
