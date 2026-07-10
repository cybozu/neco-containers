# trust-packages container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build
ENV TRUST_MANAGER_HASH=c1ba3525637adc858a8addb24b87be4d0189a77c0e532734d9aeac41b3351f71

COPY TAG /
WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d \. -f1-3 < /TAG ) \
    && curl -fsSL https://github.com/cert-manager/trust-manager/archive/v${VERSION}.tar.gz -o /tmp/trust-manager.tar.gz \
    && echo "${TRUST_MANAGER_HASH} /tmp/trust-manager.tar.gz" | sha256sum -c - \
    && tar zxf /tmp/trust-manager.tar.gz --strip-components=1 -C /work/
RUN CGO_ENABLED=0 go build -o debian-bundle-static ./trust-packages/debian/main.go

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"
LABEL description="cert-manager trust package based on Debian"

COPY --from=build /work/debian-bundle-static /debian-bundle-static
COPY --from=build /work/LICENSE /LICENSE
COPY trust-manager-source/package.json /debian-package/cert-manager-package-debian.json

USER 10000:10000

ENTRYPOINT ["/debian-bundle-static"]
CMD ["/debian-package", "/packages"]
