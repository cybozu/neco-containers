# Build stage
FROM quay.io/cybozu/golang:1.19-focal AS build

ARG SEALED_SECRETS_VERSION=0.19.5

RUN curl -sLf -o sealed-secrets.tar.gz https://github.com/bitnami-labs/sealed-secrets/archive/v${SEALED_SECRETS_VERSION}.tar.gz \
    && tar --strip-components=1 -xzf sealed-secrets.tar.gz \
    && rm sealed-secrets.tar.gz \
    && make TAG=${SEALED_SECRETS_VERSION} controller-static

# Runtime stage
FROM scratch
COPY --from=build /work/controller-static /controller
COPY --from=build /work/LICENSE /LICENSE

USER 1001:1001
EXPOSE 8080
ENTRYPOINT ["/controller"]
