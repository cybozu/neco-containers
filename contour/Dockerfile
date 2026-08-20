# Contour container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.24-noble AS build

ARG CONTOUR_VERSION=1.33.6
ARG CONTOUR_SHA=b83fb6edb8c1843d4f00f3cd653533eb88f9f843

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN git init /work/contour \
    && cd /work/contour \
    && git remote add origin https://github.com/projectcontour/contour \
    && git fetch --depth 1 origin "refs/tags/v${CONTOUR_VERSION}" \
    && test "$(git rev-parse 'FETCH_HEAD^{commit}')" = "${CONTOUR_SHA}" \
    && git checkout --detach "${CONTOUR_SHA}"

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
