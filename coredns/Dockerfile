FROM ghcr.io/cybozu/golang:1.26-noble AS build

# v1.14.2
ARG COREDNS_COMMIT=dd1df4f5db93767a44e37638df44969503b320d2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/coredns/coredns/archive/${COREDNS_COMMIT}.tar.gz | \
        tar zxf - -C /work/ \
    && mkdir -p /go/src/github.com/coredns/ \
    && mv coredns-${COREDNS_COMMIT} /go/src/github.com/coredns/coredns

WORKDIR /go/src/github.com/coredns/coredns/
RUN make

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /go/src/github.com/coredns/coredns/LICENSE /usr/local/coredns/LICENSE
COPY --from=build /go/src/github.com/coredns/coredns/coredns /usr/local/coredns/bin/coredns
ENV PATH=/usr/local/coredns/bin:"$PATH"

USER 10000:10000
WORKDIR /
EXPOSE 1053 1053/udp
ENTRYPOINT ["coredns"]
