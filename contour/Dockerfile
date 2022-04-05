# Contour container image

# Stage1: build from source
FROM quay.io/cybozu/golang:1.17-focal AS build

ARG CONTOUR_VERSION=1.20.1

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

COPY --from=build /work/contour/contour /bin/contour
COPY --from=build /work/contour/LICENSE  /LICENSE

USER 10000:10000

ENTRYPOINT ["/bin/contour"]
