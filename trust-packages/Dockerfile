# trust-packages container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.10.1_resolute@sha256:7aa0ee1eb4549f2673e0c5389ef10c3abd2ece859d48036b1b0c283415bd274e AS build
ENV TRUST_MANAGER_HASH=e4f05f49c662efa7c1df3a333d8b812f2972fbb8701352bc6607ae4120034a05

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
