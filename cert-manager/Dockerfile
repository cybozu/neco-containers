FROM ghcr.io/cybozu/golang:1.26.8.2_noble@sha256:e7d0eca1aa1de90597d2c7deeeaa7c93bc7605c51af380ccfc6776c892252a37 AS build
# https://github.com/cert-manager/cert-manager
ENV CERT_MANAGER_REPOSITORY=https://github.com/cybozu-go/cert-manager
# v1.20.4-neco.1
ENV CERT_MANAGER_SHA=a34885b2709250ca43bd27bcad5d21cc019f5dd4
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

FROM ghcr.io/cybozu/ubuntu:24.04.20260917@sha256:24c7b916b2bd28e01192045447e94c9ec4c01675131be071839bc8bffb5e218b
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/webhook /usr/local/bin/webhook
COPY --from=build /work/cainjector /usr/local/bin/cainjector
COPY --from=build /work/controller /usr/local/bin/controller
COPY --from=build /work/LICENSE /usr/local/share/doc/cert-manager/LICENSE

EXPOSE 9402

USER 10000:10000
