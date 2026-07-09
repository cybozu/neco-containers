# ExternalDNS container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG EXTERNALDNS_VERSION=0.21.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/kubernetes-sigs/external-dns/archive/v${EXTERNALDNS_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mv external-dns-${EXTERNALDNS_VERSION} /work/external-dns

WORKDIR /work/external-dns/

RUN make build

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /work/external-dns/build/external-dns /usr/local/external-dns/bin/external-dns
COPY --from=build /work/external-dns/LICENSE.md /usr/local/external-dns/LICENSE.md
ENV PATH=/usr/local/external-dns/bin:"$PATH"

USER 10000:10000

ENTRYPOINT ["external-dns"]
