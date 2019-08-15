# Gatekeeper container

# Stage1: Build gatekeeper
FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG GATEKEEPER_VERSION=3.0.4-beta.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/open-policy-agent/gatekeeper/archive/v${GATEKEEPER_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mkdir -p /go/src/github.com/open-policy-agent \
    && mv gatekeeper-${GATEKEEPER_VERSION} /go/src/github.com/open-policy-agent/gatekeeper

WORKDIR /go/src/github.com/open-policy-agent/gatekeeper/

# Build
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/open-policy-agent/gatekeeper/cmd/manager


# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

EXPOSE 8443
USER 10000:10000

WORKDIR /
COPY --from=build /go/src/github.com/open-policy-agent/gatekeeper/manager /usr/local/gatekeeper/bin/manager
COPY --from=build /go/src/github.com/open-policy-agent/gatekeeper/LICENSE /usr/local/gatekeeper/bin/LICENSE

ENV PATH=/usr/local/gatekeeper/bin:"$PATH"
ENTRYPOINT ["manager"]
