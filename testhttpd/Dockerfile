# testhttpd container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build
COPY src /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./testhttpd

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin /
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/testhttpd", "-listen", ":8000"]
