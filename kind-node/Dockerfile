# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build
COPY src /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -mod=vendor ./ptp

# Stage2: setup runtime container
FROM kindest/node:v1.16.3
COPY --from=build /go/bin/ptp /opt/cni/bin/ptp
