# stakater/Reloader container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG RELOADER_SHA=cff4592fbdc942aba17d205860818e8ce45e6696

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
