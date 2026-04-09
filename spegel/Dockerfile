FROM ghcr.io/cybozu/golang:1.25-noble AS builder
ARG SPEGEL_COMMIT=6f02215fa3fc1d3bbdb11fa62dfa7c07dbe3b7c2

WORKDIR /work/spegel

RUN curl -sSLf https://github.com/spegel-org/spegel/archive/${SPEGEL_COMMIT}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/spegel

RUN go mod download
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o spegel .

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=builder /work/spegel /app/
WORKDIR /app

ENTRYPOINT ["./spegel"]
