# cpu-lending-agent container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.25-noble AS build
COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o cpu-lending-agent

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/cpu-lending-agent /
EXPOSE 8080/tcp
ENTRYPOINT ["/cpu-lending-agent"]
