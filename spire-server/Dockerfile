# spire-server container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.25-noble AS build

COPY TAG /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d . -f 1,2,3 < /TAG) \
    && git clone --depth=1 -b v${VERSION} https://github.com/spiffe/spire.git

WORKDIR /work/spire

# CGO is required for mattn/go-sqlite3 compatibility
ENV CGO_ENABLED=1
RUN make build-static BINARY=spire-server && \
    mv bin/static/spire-server /work/spire-server && \
    (ldd /work/spire-server 2>&1 || true) | grep -q "not a dynamic executable"

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /work/spire-server /usr/local/bin/spire-server
COPY --from=build /work/spire/LICENSE /usr/local/share/doc/spire-server/LICENSE

USER 10000:10000
EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/spire-server"]
