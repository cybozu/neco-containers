# trust-manager container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.20-jammy AS build

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

WORKDIR /
COPY --from=build /workspace/manager .
COPY --from=build /workspace/LICENSE /LICENSE

USER 10000:10000
EXPOSE 9090

ENTRYPOINT ["/manager"]
