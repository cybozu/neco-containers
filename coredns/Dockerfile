FROM quay.io/cybozu/golang:1.19-focal AS build

ARG COREDNS_VERSION=1.10.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/coredns/coredns/archive/v${COREDNS_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mkdir -p /go/src/github.com/coredns/ \
    && mv coredns-${COREDNS_VERSION} /go/src/github.com/coredns/coredns

WORKDIR /go/src/github.com/coredns/coredns/
RUN make

FROM quay.io/cybozu/ubuntu:20.04

COPY --from=build /go/src/github.com/coredns/coredns/LICENSE /usr/local/coredns/LICENSE
COPY --from=build /go/src/github.com/coredns/coredns/coredns /usr/local/coredns/bin/coredns
ENV PATH=/usr/local/coredns/bin:"$PATH"

USER 10000:10000
EXPOSE 1053 1053/udp
ENTRYPOINT ["coredns"]
