FROM ghcr.io/cybozu/golang:1.27-resolute AS builder
ENV GOPROXY=https://golang.flatt.tech
ARG SPEGEL_COMMIT=7e2e042fa5a6858276b12bbd0be586b788dfa121 # v0.8.1

WORKDIR /work/spegel

RUN curl -sSLf https://github.com/spegel-org/spegel/archive/${SPEGEL_COMMIT}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/spegel

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    go mod download
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o spegel .

# - Spegel checks its runtime environment with a preflight check.
#   Since v0.7.4, it fails to start unless it runs on a distroless image.
#   https://github.com/spegel-org/spegel/pull/1448
# - The upstream image uses `static:nonroot` but switches to root with `USER root:root`.
#   We use `static:latest` instead, which runs as root by default.
#   https://github.com/spegel-org/spegel/blob/v0.8.1/Dockerfile#L5
FROM gcr.io/distroless/static:latest@sha256:1a64f46364620465ef9e9a8f5413757dfdf22695e050bbfd7303eeaf17a59383
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/spegel /app/
WORKDIR /app

ENTRYPOINT ["./spegel"]
