# syntax=docker/dockerfile:1
# https://github.com/argoproj/argo-rollouts/blob/master/Dockerfile
# Remove following steps from original Dockerfile.
# - UI build stage
# - Kubectl plugin image

FROM scratch AS argo-rollouts-src
ARG ARGO_ROLLOUTS_VERSION=1.10.0
ADD --checksum=sha256:843109c33b1fea73475c5d56255ea00149663964c5c02382832a1dcf51d5219a \
    https://github.com/argoproj/argo-rollouts/archive/refs/tags/v${ARGO_ROLLOUTS_VERSION}.tar.gz /src.tar.gz

FROM ghcr.io/cybozu/golang:1.26.10.2_noble@sha256:8d900a9ee78c02dbe4e2d2e23800c4d21714dfaf8c679935afa668ca0993a0e4 AS builder
ENV GOPROXY=https://golang.flatt.tech

WORKDIR /work/argo-rollouts

RUN --mount=type=bind,from=argo-rollouts-src,source=/src.tar.gz,target=/tmp/src.tar.gz \
    tar xzf /tmp/src.tar.gz --strip-components=1 -C .

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    go mod download

RUN make controller

FROM gcr.io/distroless/static-debian12@sha256:d75cdd72874d4790092fcb1b058493ecf6bb5bf2b2b897045b00ff01d91843f2
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/argo-rollouts/dist/rollouts-controller /bin/
COPY --from=builder /work/argo-rollouts/LICENSE /LICENSE
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 10000:10000

WORKDIR /home/argo-rollouts

ENTRYPOINT [ "/bin/rollouts-controller" ]
