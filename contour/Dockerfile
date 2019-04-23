FROM quay.io/cybozu/golang:1.12-bionic AS build

ARG CONTOUR_VERSION=0.11.0

RUN curl -sSLf https://github.com/heptio/contour/archive/v${CONTOUR_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mv contour-${CONTOUR_VERSION} /work/contour

WORKDIR /work/contour/

ENV GOPROXY=https://gocenter.io
RUN go mod download
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOFLAGS=-ldflags=-w go build -o /go/bin/contour -ldflags=-s -v github.com/heptio/contour/cmd/contour

FROM scratch

COPY --from=build /go/bin/contour /usr/local/contour/bin/contour
COPY --from=build /work/contour/LICENSE  /usr/local/contour/LICENSE
ENV PATH=/usr/local/contour/bin:"$PATH"

USER 10000:10000

ENTRYPOINT ["contour"]
