# trust-manager container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.21-jammy AS build

COPY TAG /
WORKDIR /workspace
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && curl -fsSL https://github.com/stakater/Reloader/archive/v${VERSION}.tar.gz | \
    tar --strip-components=1 -xzf -

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod=mod -a -o manager main.go

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

WORKDIR /
COPY --from=build /workspace/manager .
COPY --from=build /workspace/LICENSE /LICENSE

USER 10000:10000
EXPOSE 9090

ENTRYPOINT ["/manager"]
