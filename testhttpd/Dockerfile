# testhttpd container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.17-focal AS build
COPY src /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go install -ldflags="-w -s" ./testhttpd

# Stage2: setup runtime container
FROM scratch
COPY --from=build /go/bin /
USER 10000:10000
EXPOSE 8000
ENTRYPOINT ["/testhttpd", "-listen", ":8000"]
