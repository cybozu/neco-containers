# ExternalDNS container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27.3.1_resolute@sha256:56354d624535147898457543f53b6b42bebd671817e0ec322b610b1370c98f2c AS build
ENV EXTERNALDNS_HASH=f9e61848ab12b9bfbd19e64a8a4c01085534f1ba82b90c46276ff173909dd3d5

ARG EXTERNALDNS_VERSION=0.23.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/kubernetes-sigs/external-dns/archive/v${EXTERNALDNS_VERSION}.tar.gz -o /tmp/external-dns.tar.gz \
    && echo "${EXTERNALDNS_HASH} /tmp/external-dns.tar.gz" | sha256sum -c - \
    && mkdir -p /work/external-dns \
    && tar zxf /tmp/external-dns.tar.gz --strip-components=1 -C /work/external-dns/

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
