FROM ghcr.io/cybozu/golang:1.27.3.2_resolute@sha256:2e42484cb145c45c51065e58a62420b3afa2164ed303062ceb2f7b82e978df20 AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o cep-checker

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/cep-checker /
EXPOSE 8080/tcp
ENTRYPOINT ["/cep-checker"]
