FROM ghcr.io/cybozu/golang:1.26.4.1_noble@sha256:add9d704d4b75df2c51328615be89b61a3e71e4833321aa02c3f325a30d3eb8f AS build
COPY . /work
RUN CGO_ENABLED=0 go build -o bin/neco-admission -ldflags="-w -s" main.go

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/bin/neco-admission /neco-admission
USER 10000:10000
ENTRYPOINT ["/neco-admission"]
