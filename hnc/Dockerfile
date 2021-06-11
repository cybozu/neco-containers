FROM quay.io/cybozu/golang:1.16-focal AS build
ARG HNC_VERSION=0.8.0

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone --depth=1 -b hnc-v${HNC_VERSION} https://github.com/kubernetes-sigs/hierarchical-namespaces /work/hnc
RUN cd hnc/incubator/hnc && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager ./cmd/manager/main.go

FROM scratch

COPY --from=build /work/hnc/incubator/hnc/manager /manager

USER 10000:10000
ENTRYPOINT ["/manager"]
