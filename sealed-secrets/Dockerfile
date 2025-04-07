# Build stage
FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG SEALED_SECRETS_VERSION=0.29.0

RUN curl -sLf -o sealed-secrets.tar.gz https://github.com/bitnami-labs/sealed-secrets/archive/v${SEALED_SECRETS_VERSION}.tar.gz \
    && tar --strip-components=1 -xzf sealed-secrets.tar.gz \
    && rm sealed-secrets.tar.gz \
    && make TAG=${SEALED_SECRETS_VERSION} controller-static

# Runtime stage
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/controller-static /controller
COPY --from=build /work/LICENSE /LICENSE

USER 1001:1001
EXPOSE 8080 8081
ENTRYPOINT ["/controller"]
