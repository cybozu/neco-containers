# stakater/Reloader container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.3.2_resolute@sha256:2e42484cb145c45c51065e58a62420b3afa2164ed303062ceb2f7b82e978df20 AS build

ARG RELOADER_SHA=93bac9d625556bdbd7323616fe1b05072072cdda

COPY TAG /
WORKDIR /workspace
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d \. -f 1,2,3 < /TAG ) \
    && git init \
    && git remote add origin https://github.com/stakater/Reloader \
    && git fetch --depth 1 origin "refs/tags/v${VERSION}" \
    && test "$(git rev-parse 'FETCH_HEAD^{commit}')" = "${RELOADER_SHA}" \
    && git checkout --detach "${RELOADER_SHA}"

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
