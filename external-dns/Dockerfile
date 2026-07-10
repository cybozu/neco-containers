# ExternalDNS container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build
ENV EXTERNALDNS_HASH=78fd72654681bb6ebf13f1360551f8aafb08f2d0eeecdd6523a89c8631b5cf64

ARG EXTERNALDNS_VERSION=0.21.0

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
