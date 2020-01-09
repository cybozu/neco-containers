# local-pv-provisioner container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 GO111MODULE=on go install -mod=vendor -ldflags="-w -s" .

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/bin /

USER 10000:10000

ENTRYPOINT ["/local-pv-provisioner"]
