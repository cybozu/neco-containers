# metallb container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG METALLB_VERSION=0.7.3

WORKDIR /go/src/go.universe.tf/metallb

RUN curl -fsSL -o metallb.tar.gz https://github.com/google/metallb/archive/v${METALLB_VERSION}.tar.gz \
    && tar -x -z --strip-components 1 -f metallb.tar.gz \
    && rm -f metallb.tar.gz \
    && CGO_ENABLED=0 GOLDFLAGS="-w -s" go install ./speaker ./controller

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/bin/speaker /speaker
COPY --from=build /go/bin/controller /controller
COPY --from=build /go/src/go.universe.tf/metallb/LICENSE /LICENSE

ENTRYPOINT ["/speaker"]
