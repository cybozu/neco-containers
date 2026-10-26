FROM ghcr.io/cybozu/golang:1.27.3.2_resolute@sha256:2e42484cb145c45c51065e58a62420b3afa2164ed303062ceb2f7b82e978df20 AS build
# https://github.com/cert-manager/cert-manager
ENV CERT_MANAGER_REPOSITORY=https://github.com/cybozu-go/cert-manager
# v1.21.3-neco.1
ENV CERT_MANAGER_SHA=ff37b908adbbff4647cc3c822f8d380156b05a84
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

FROM gcr.io/distroless/static-debian13:latest@sha256:15f8637509792e05b97d83451b0b6ccfa4abd548f997e406ac1e72c87fe1fa78
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/webhook /usr/local/bin/webhook
COPY --from=build /work/cainjector /usr/local/bin/cainjector
COPY --from=build /work/controller /usr/local/bin/controller
COPY --from=build /work/LICENSE /usr/local/share/doc/cert-manager/LICENSE

EXPOSE 9402

USER 10000:10000
