FROM ghcr.io/cybozu/golang:1.26.10.1_resolute@sha256:7aa0ee1eb4549f2673e0c5389ef10c3abd2ece859d48036b1b0c283415bd274e AS build
# https://github.com/cert-manager/cert-manager
ENV CERT_MANAGER_REPOSITORY=https://github.com/cybozu-go/cert-manager
# v1.21.2-neco.1
ENV CERT_MANAGER_SHA=aadb4a01f7cbe0b249154b6e71a6665689554fc7
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

FROM gcr.io/distroless/static-debian13:latest@sha256:9e0130bec991f471279b70f1630ada11e46dad439ef2949544f96001225329fb
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/webhook /usr/local/bin/webhook
COPY --from=build /work/cainjector /usr/local/bin/cainjector
COPY --from=build /work/controller /usr/local/bin/controller
COPY --from=build /work/LICENSE /usr/local/share/doc/cert-manager/LICENSE

EXPOSE 9402

USER 10000:10000
