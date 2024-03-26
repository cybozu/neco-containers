# ttypdb-xxx container

FROM ghcr.io/cybozu/golang:1.22-jammy AS build

COPY . /work
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./...


FROM scratch AS sidecar
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/sidecar /ttypdb-sidecar
USER 10000:10000
ENTRYPOINT ["/ttypdb-sidecar"]



FROM scratch AS controller
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/controller /ttypdb-controller
USER 10000:10000
ENTRYPOINT ["/ttypdb-controller"]
