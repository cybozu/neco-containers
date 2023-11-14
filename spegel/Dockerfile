FROM ghcr.io/cybozu/golang:1.21-jammy AS builder
ARG SPEGEL_VERSION=0.0.14

WORKDIR /work/spegel

RUN curl -sSLf https://github.com/XenitAB/spegel/archive/v${SPEGEL_VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/spegel

RUN go mod download
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o spegel .

FROM scratch
COPY --from=builder /work/spegel /app/
WORKDIR /app

ENTRYPOINT ["./spegel"]
