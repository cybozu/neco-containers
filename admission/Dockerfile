FROM ghcr.io/cybozu/golang:1.27.4.2_resolute@sha256:0ddbebb97d9fefa0961dd10fdccedc40a4b805d28f21f21d92faa2f7b071d2f1 AS build
ENV GOPROXY=https://golang.flatt.tech
COPY . /work
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    CGO_ENABLED=0 go build -o bin/neco-admission -ldflags="-w -s" ./cmd

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/bin/neco-admission /neco-admission
USER 10000:10000
ENTRYPOINT ["/neco-admission"]
