FROM ghcr.io/cybozu/golang:1.24-noble AS build
# https://github.com/cert-manager/cert-manager
ENV CERT_MANAGER_REPOSITORY=https://github.com/cybozu-go/cert-manager
ENV CERT_MANAGER_VERSION=1.17.3-neco-longtimeout.1
RUN git clone --depth 1 --branch v${CERT_MANAGER_VERSION} ${CERT_MANAGER_REPOSITORY}
RUN apt update && apt-get install -y rsync
RUN cd cert-manager \
    && make CTR=sleep _bin/server/controller-linux-amd64 \
    && make CTR=sleep _bin/server/webhook-linux-amd64 \
    && make CTR=sleep _bin/server/cainjector-linux-amd64 \
    && mv _bin/server/controller-linux-amd64 /work/controller \
    && mv _bin/server/webhook-linux-amd64 /work/webhook \
    && mv _bin/server/cainjector-linux-amd64 /work/cainjector \
    && mv LICENSE /work

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/webhook /usr/local/bin/webhook
COPY --from=build /work/cainjector /usr/local/bin/cainjector
COPY --from=build /work/controller /usr/local/bin/controller
COPY --from=build /work/LICENSE /usr/local/share/doc/cert-manager/LICENSE

EXPOSE 9402

USER 10000:10000
