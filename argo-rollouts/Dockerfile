# syntax=docker/dockerfile:1
# https://github.com/argoproj/argo-rollouts/blob/master/Dockerfile
# Remove following steps from original Dockerfile.
# - UI build stage
# - Kubectl plugin image

FROM scratch AS argo-rollouts-src
ARG ARGO_ROLLOUTS_VERSION=1.10.0
ADD --checksum=sha256:843109c33b1fea73475c5d56255ea00149663964c5c02382832a1dcf51d5219a \
    https://github.com/argoproj/argo-rollouts/archive/refs/tags/v${ARGO_ROLLOUTS_VERSION}.tar.gz /src.tar.gz

FROM ghcr.io/cybozu/golang:1.26.8.2_noble@sha256:e7d0eca1aa1de90597d2c7deeeaa7c93bc7605c51af380ccfc6776c892252a37 AS builder

WORKDIR /work/argo-rollouts

RUN --mount=type=bind,from=argo-rollouts-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .

RUN go mod download

RUN make controller

FROM gcr.io/distroless/static-debian12@sha256:d75cdd72874d4790092fcb1b058493ecf6bb5bf2b2b897045b00ff01d91843f2
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/argo-rollouts/dist/rollouts-controller /bin/
COPY --from=builder /work/argo-rollouts/LICENSE /LICENSE
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 10000:10000

WORKDIR /home/argo-rollouts

ENTRYPOINT [ "/bin/rollouts-controller" ]
