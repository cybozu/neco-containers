FROM ghcr.io/cybozu/golang:1.22-jammy AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o cep-checker

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/cep-checker /
EXPOSE 8080/tcp
ENTRYPOINT ["/cep-checker"]
