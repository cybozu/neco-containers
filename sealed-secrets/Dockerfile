# Build stage
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG SEALED_SECRETS_VERSION=0.38.1
ARG SEALED_SECRETS_COMMIT=706528904566e28373b584b5adb7e92ef8f81c1d

RUN curl -sLf -o sealed-secrets.tar.gz https://github.com/bitnami/sealed-secrets/archive/${SEALED_SECRETS_COMMIT}.tar.gz \
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
