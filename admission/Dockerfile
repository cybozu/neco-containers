FROM ghcr.io/cybozu/golang:1.23-noble AS build
COPY . /work
RUN CGO_ENABLED=0 go build -o bin/neco-admission -ldflags="-w -s" main.go

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/bin/neco-admission /neco-admission
USER 10000:10000
ENTRYPOINT ["/neco-admission"]
