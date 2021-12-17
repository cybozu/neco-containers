# metallb container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.17-focal AS build

ARG METALLB_VERSION=0.11.0

WORKDIR /work/metallb

RUN curl -fsSL -o metallb.tar.gz https://github.com/metallb/metallb/archive/v${METALLB_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f metallb.tar.gz \
    && rm -f metallb.tar.gz \
    && CGO_ENABLED=0 GOLDFLAGS="-w -s" go install ./speaker ./controller

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/bin/speaker /speaker
COPY --from=build /go/bin/controller /controller
COPY --from=build /work/metallb/LICENSE /LICENSE

USER 10000:10000

ENTRYPOINT ["/speaker"]
