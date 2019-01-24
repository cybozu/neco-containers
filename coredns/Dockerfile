FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG COREDNS_VERSION=1.3.1
RUN update-ca-certificates

RUN curl -sSLf https://github.com/coredns/coredns/archive/v${COREDNS_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mkdir -p /go/src/github.com/coredns/ \
    && mv coredns-${COREDNS_VERSION} /go/src/github.com/coredns/coredns

WORKDIR /go/src/github.com/coredns/coredns/
RUN make 

FROM scratch

COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /go/src/github.com/coredns/coredns/LICENSE /usr/local/coredns/LICENSE
COPY --from=build /go/src/github.com/coredns/coredns/coredns /usr/local/coredns/bin/coredns
ENV PATH=/usr/local/coredns/bin:"$PATH"

EXPOSE 53 53/udp
ENTRYPOINT ["coredns"]
