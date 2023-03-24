# s3gw container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.20-jammy AS build

COPY . /work

WORKDIR /work

RUN CGO_ENABLED=0 go install -ldflags="-w -s" .

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/bin /

USER 10000:10000

ENTRYPOINT ["/s3gw"]
