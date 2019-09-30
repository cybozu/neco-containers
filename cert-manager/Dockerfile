# cert-manager container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build
ARG CERT_MANAGER_VERSION=0.10.1

WORKDIR /go/src/github.com/jetstack/cert-manager

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/jetstack/cert-manager/archive/v${CERT_MANAGER_VERSION}.tar.gz | tar zxf - --strip-components 1 && \
    GO111MODULE=on go install -mod=vendor ./cmd/webhook/ ./cmd/cainjector ./cmd/controller

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin /usr/local/cert-manager/bin
COPY --from=build /go/src/github.com/jetstack/cert-manager/LICENSE /usr/local/cert-manager/LICENSE

EXPOSE 9402

USER 10000:10000

ENV PATH=/usr/local/cert-manager/bin:"$PATH"
