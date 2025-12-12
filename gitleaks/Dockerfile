FROM ghcr.io/cybozu/golang:1.25-noble AS build
ARG GITLEAKS_VERSION=8.30.0

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN git clone --depth=1 -b v${GITLEAKS_VERSION} https://github.com/gitleaks/gitleaks /work/gitleaks

WORKDIR /work/gitleaks

RUN CGO_ENABLED=0 go build -ldflags '-w -s' -o bin/gitleaks

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

WORKDIR /
COPY --from=build /work/gitleaks/bin/gitleaks /usr/bin/gitleaks
COPY --from=build /work/gitleaks/LICENSE /LICENSE

USER 10000:10000
ENTRYPOINT ["/usr/bin/gitleaks"]
