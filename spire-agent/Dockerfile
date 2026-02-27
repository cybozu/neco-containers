# spire-agent container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.25-noble AS build

COPY TAG /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d . -f 1,2,3 < /TAG) \
    && git clone --depth=1 -b v${VERSION} https://github.com/spiffe/spire.git

WORKDIR /work/spire

# CGO is required for mattn/go-sqlite3 compatibility
ENV CGO_ENABLED=1
RUN make build-static BINARY=spire-agent && \
    mv bin/static/spire-agent /work/spire-agent && \
    (ldd /work/spire-agent 2>&1 || true) | grep -q "not a dynamic executable"

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /work/spire-agent /usr/local/bin/spire-agent
COPY --from=build /work/spire/LICENSE /usr/local/share/doc/spire-agent/LICENSE

ENTRYPOINT ["/usr/local/bin/spire-agent", "run"]
