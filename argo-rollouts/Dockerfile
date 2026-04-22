# syntax=docker/dockerfile:1
# https://github.com/argoproj/argo-rollouts/blob/master/Dockerfile
# Remove following steps from original Dockerfile.
# - UI build stage
# - Kubectl plugin image

FROM scratch AS argo-rollouts-src
ARG ARGO_ROLLOUTS_VERSION=1.9.0
ADD --checksum=sha256:21255db36501d753bd1a20c686613cd225426c545a9471000f303499f8d84163 \
    https://github.com/argoproj/argo-rollouts/archive/refs/tags/v${ARGO_ROLLOUTS_VERSION}.tar.gz /src.tar.gz

FROM ghcr.io/cybozu/golang:1.26.2.2_noble@sha256:8154343edf8b656f4851572321483b8f304fa00b9966f1db687a7c1a3f234da8 AS builder

WORKDIR /work/argo-rollouts

RUN --mount=type=bind,from=argo-rollouts-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .

RUN go mod download

RUN make controller

FROM gcr.io/distroless/static-debian12@sha256:20bc6c0bc4d625a22a8fde3e55f6515709b32055ef8fb9cfbddaa06d1760f838
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/argo-rollouts/dist/rollouts-controller /bin/
COPY --from=builder /work/argo-rollouts/LICENSE /LICENSE
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 10000:10000

WORKDIR /home/argo-rollouts

ENTRYPOINT [ "/bin/rollouts-controller" ]
