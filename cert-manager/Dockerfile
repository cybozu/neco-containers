FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build
# https://github.com/cert-manager/cert-manager
ENV CERT_MANAGER_REPOSITORY=https://github.com/cybozu-go/cert-manager
# v1.20.3-neco.1
ENV CERT_MANAGER_SHA=cfb385d376968356eeb793162d664380771f39ba
RUN mkdir cert-manager \
    && cd cert-manager \
    && git init 2>/dev/null \
    && git remote add origin ${CERT_MANAGER_REPOSITORY} \
    && git fetch --depth 1 origin ${CERT_MANAGER_SHA} \
    && git checkout FETCH_HEAD 2>/dev/null
RUN apt update && apt-get install -y rsync
RUN cd cert-manager \
    && make CTR=sleep _bin/server/controller-linux-amd64 \
    && make CTR=sleep _bin/server/webhook-linux-amd64 \
    && make CTR=sleep _bin/server/cainjector-linux-amd64 \
    && mv _bin/server/controller-linux-amd64 /work/controller \
    && mv _bin/server/webhook-linux-amd64 /work/webhook \
    && mv _bin/server/cainjector-linux-amd64 /work/cainjector \
    && mv LICENSE /work

FROM ghcr.io/cybozu/ubuntu:24.04.20260709@sha256:f91f1884f5b8dbe44e224306b5dc7bfd4d420f5c73aea1d32a71c6474361fb0f
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/webhook /usr/local/bin/webhook
COPY --from=build /work/cainjector /usr/local/bin/cainjector
COPY --from=build /work/controller /usr/local/bin/controller
COPY --from=build /work/LICENSE /usr/local/share/doc/cert-manager/LICENSE

EXPOSE 9402

USER 10000:10000
