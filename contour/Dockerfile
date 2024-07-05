# Contour container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG CONTOUR_VERSION=1.29.1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/projectcontour/contour/archive/v${CONTOUR_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mv contour-${CONTOUR_VERSION} /work/contour

WORKDIR /work/contour/

RUN make build \
    CGO_ENABLED=0 \
    GOOS=linux

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/contour/contour /bin/contour
COPY --from=build /work/contour/LICENSE  /LICENSE

USER 10000:10000

ENTRYPOINT ["/bin/contour"]
