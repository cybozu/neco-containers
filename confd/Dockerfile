FROM quay.io/cybozu/golang:1.11-bionic AS build

ARG CONFD_VERSION=0.16.0

RUN git clone https://github.com/kelseyhightower/confd /go/src/github.com/kelseyhightower/confd

WORKDIR /go/src/github.com/kelseyhightower/confd
RUN git checkout tags/v${CONFD_VERSION} \
    && go get -u github.com/golang/dep/cmd/dep \
    && make build

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/src/github.com/kelseyhightower/confd/bin/confd /usr/local/confd/bin/confd
COPY --from=build /go/src/github.com/kelseyhightower/confd/LICENSE /usr/local/confd/LICENSE
ENV PATH=/usr/local/confd/bin:"$PATH"

ENTRYPOINT ["confd"]
