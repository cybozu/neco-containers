# httpd container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.12-bionic AS build
COPY src /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -mod=vendor ./httpd

# Stage2: setup runtime container
FROM scratch
COPY --from=build /go/bin /
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/httpd", "-p", "8000"]
