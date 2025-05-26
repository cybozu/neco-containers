# https://github.com/argoproj/argo-rollouts/blob/master/Dockerfile
# Remove following steps from original Dockerfile.
# - UI build stage
# - Kubectl plugin image

FROM ghcr.io/cybozu/golang:1.23-noble AS builder

ENV ARGO_ROLLOUTS_VERSION=1.8.2

WORKDIR /work

RUN git clone --depth 1 --branch v${ARGO_ROLLOUTS_VERSION} https://github.com/argoproj/argo-rollouts.git

WORKDIR /work/argo-rollouts

RUN go mod download

RUN make controller

FROM gcr.io/distroless/static-debian11
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/argo-rollouts/dist/rollouts-controller /bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 10000:10000

WORKDIR /home/argo-rollouts

ENTRYPOINT [ "/bin/rollouts-controller" ]
