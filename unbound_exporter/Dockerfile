# unbound_exporter container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build

#ARG UNBOUND_EXPORTER_REPO=letsencrypt/unbound_exporter
ARG UNBOUND_EXPORTER_REPO=cybozu-go/unbound_exporter
ARG UNBOUND_EXPORTER_VERSION=0.4.6-neco.1

WORKDIR /work

RUN curl -fsSL -o unbound_exporter.tar.gz "https://github.com/${UNBOUND_EXPORTER_REPO}/archive/v${UNBOUND_EXPORTER_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f unbound_exporter.tar.gz \
    && rm -f unbound_exporter.tar.gz \
    && CGO_ENABLED=0 go install -ldflags="-w -s" .

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/bin/unbound_exporter /bin/unbound_exporter
COPY --from=build /work/LICENSE /

USER 10000:10000

ENTRYPOINT ["/bin/unbound_exporter"]
