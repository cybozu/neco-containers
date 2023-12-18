FROM ghcr.io/cybozu/golang:1.20-jammy AS build
COPY . /work
RUN make build

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/bin/neco-admission /neco-admission
USER 10000:10000
ENTRYPOINT ["/neco-admission"]
