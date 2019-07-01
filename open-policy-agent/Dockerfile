# Open Policy Agent container

# Stage1: Initial Stage which pulls prepares build dependencies and CLI tooling we need for out final image
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG OPA_VERSION=0.12.0
ARG KUBE_MGMT_VERSION=0.8
ARG KUBE_MGMT_PKG=github.com/open-policy-agent/kube-mgmt

# Install Open Policy Agent
RUN curl -sSLf -o /usr/local/bin/opa https://github.com/open-policy-agent/opa/releases/download/v${OPA_VERSION}/opa_linux_amd64 && \
    chmod +x /usr/local/bin/opa && \
    curl -sSLf -o /usr/local/bin/LICENSE-opa https://raw.githubusercontent.com/open-policy-agent/opa/v0.12.0/LICENSE

# Install kube-mgmt sidecar container
WORKDIR /go/src/${KUBE_MGMT_PKG}
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/open-policy-agent/kube-mgmt/archive/v${KUBE_MGMT_VERSION}.tar.gz | \
        tar -x -z -f - --strip-components 1 && \
    GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -mod=vendor -ldflags="-w -s -X ${KUBE_MGMT_PKG}/pkg/version.Version=${KUBE_MGMT_VERSION}" -v ./cmd/kube-mgmt/... && \
    cp LICENSE /go/bin/LICENSE-kube-mgmt

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /usr/local/bin/opa /usr/local/open-policy-agent/bin/opa
COPY --from=build /usr/local/bin/LICENSE-opa /usr/local/open-policy-agent/bin/LICENSE-opa
COPY --from=build /go/bin/kube-mgmt /usr/local/open-policy-agent/bin/kube-mgmt
COPY --from=build /go/bin/LICENSE-kube-mgmt /usr/local/open-policy-agent/bin/LICENSE-kube-mgmt

EXPOSE 8443

USER 10000:10000

ENV PATH=/usr/local/open-policy-agent/bin:"$PATH"

ENTRYPOINT ["opa"]
