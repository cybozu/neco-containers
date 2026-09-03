FROM ghcr.io/cybozu/golang:1.26.8.1_noble@sha256:fcdd7870a523950129a6afc09e03c035796446ac7c1b2f84994e4a7920042879 AS build
COPY . /work
RUN CGO_ENABLED=0 go build -o bin/neco-admission -ldflags="-w -s" main.go

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/bin/neco-admission /neco-admission
USER 10000:10000
ENTRYPOINT ["/neco-admission"]
