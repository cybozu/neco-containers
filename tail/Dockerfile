# tail container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.13-bionic AS build
COPY src /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -mod=vendor ./tail

# Stage2: setup runtime container
FROM scratch
COPY --from=build /go/bin /
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/tail"]
