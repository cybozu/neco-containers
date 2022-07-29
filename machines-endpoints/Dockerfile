# machines-endpoints container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.18-focal AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./pkg/machines-endpoints

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/bin /

USER 10000:10000
EXPOSE 7373 7946

ENTRYPOINT ["/machines-endpoints"]
