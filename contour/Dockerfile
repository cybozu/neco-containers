# Contour container image

# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build

ARG CONTOUR_VERSION=1.4.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/projectcontour/contour/archive/v${CONTOUR_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mv contour-${CONTOUR_VERSION} /work/contour

WORKDIR /work/contour/

RUN make install \
    CGO_ENABLED=0 \
    GOOS=linux

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin/contour /bin/contour
COPY --from=build /work/contour/LICENSE  /usr/local/contour/LICENSE

USER 10000:10000

ENTRYPOINT ["contour"]
