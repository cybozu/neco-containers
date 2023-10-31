# trust-packages container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.21-jammy AS build

ARG TRUST_MANAGER_VERSION=0.7.0
WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -fsSL https://github.com/cert-manager/trust-manager/archive/v${TRUST_MANAGER_VERSION}.tar.gz | \
    tar --strip-components=1 -xzf -

WORKDIR /work/trust-packages/debian
RUN CGO_ENABLED=0 go build -o copyandmaybepause main.go
RUN CGO_ENABLED=0 go install github.com/cert-manager/trust-manager/cmd/validate-trust-package@main

# Stage2: copy ca-certificates from ubuntu
FROM quay.io/cybozu/ubuntu:22.04 as debbase

ARG EXPECTED_VERSION=20230311ubuntu0.22.04.1
ARG VERSION_SUFFIX=.0
ARG TINI_VERSION=0.19.0
WORKDIR /work

COPY --from=build /work/trust-packages/debian/build.sh /work/build.sh
COPY --from=build /go/bin/validate-trust-package /usr/bin/validate-trust-package

RUN /work/build.sh $EXPECTED_VERSION $VERSION_SUFFIX /work/package.json
RUN curl -fsSL -o /usr/bin/tini-static https://github.com/krallin/tini/releases/download/v${TINI_VERSION}/tini-static

# Stage3: setup runtime container
FROM scratch

LABEL description="cert-manager trust package based on Debian"

COPY --from=debbase /usr/bin/tini-static /tini
COPY --from=debbase /work/package.json /debian-package/cert-manager-package-debian.json
COPY --from=build /work/trust-packages/debian/copyandmaybepause /copyandmaybepause
COPY --from=build /work/LICENSE /LICENSE

ENTRYPOINT ["/tini", "--"]

USER 10000:10000
CMD ["/copyandmaybepause", "/debian-package", "/packages"]
