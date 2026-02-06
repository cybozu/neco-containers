# spire-controller-manager container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.25-noble AS build

COPY TAG /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d . -f 1,2,3 < /TAG) \
    && git clone --depth=1 -b v${VERSION} https://github.com/spiffe/spire-controller-manager.git

WORKDIR /work/spire-controller-manager

ENV CGO_ENABLED=0
RUN go build -o /work/spire-controller-manager-bin cmd/main.go

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /work/spire-controller-manager-bin /usr/local/bin/spire-controller-manager
COPY --from=build /work/spire-controller-manager/LICENSE /usr/local/share/doc/spire-controller-manager/LICENSE

USER 10000:10000

ENTRYPOINT ["/usr/local/bin/spire-controller-manager"]
