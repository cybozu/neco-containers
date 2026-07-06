FROM ghcr.io/cybozu/golang:1.26-noble AS build
ARG GITLEAKSS_COMMIT=83d9cd684c87d95d656c1458ef04895a7f1cbd8e # 8.30.1

WORKDIR /work/gitleaks
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN curl -sLf -o gitleaks.tar.gz https://github.com/gitleaks/gitleaks/archive/${GITLEAKSS_COMMIT}.tar.gz \
    && tar --strip-components=1 -xzf gitleaks.tar.gz -C /work/gitleaks \
    && rm gitleaks.tar.gz

RUN CGO_ENABLED=0 go build -ldflags '-w -s' -o bin/gitleaks

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

WORKDIR /
COPY --from=build /work/gitleaks/bin/gitleaks /usr/bin/gitleaks
COPY --from=build /work/gitleaks/LICENSE /LICENSE

USER 10000:10000
ENTRYPOINT ["/usr/bin/gitleaks"]
