FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o cep-checker

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/cep-checker /
EXPOSE 8080/tcp
ENTRYPOINT ["/cep-checker"]
