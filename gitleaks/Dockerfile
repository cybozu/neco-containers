FROM ghcr.io/cybozu/golang:1.25-noble AS build
ARG GITLEAKSS_COMMIT=6eaad039603a4de39fddd1cf5f727391efe9974e

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
