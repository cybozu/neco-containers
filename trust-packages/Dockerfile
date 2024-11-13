# trust-packages container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.23-jammy AS build

ARG TRUST_MANAGER_VERSION=0.13.0
WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/cert-manager/trust-manager/archive/v${TRUST_MANAGER_VERSION}.tar.gz | \
    tar --strip-components=1 -xzf -
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
