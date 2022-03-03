# cadvisor container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.17-focal AS build

ARG CADVISOR_VERSION=0.44.0

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /go/src/github.com/google/cadvisor
RUN curl -fsSL -o cadvisor.tar.gz "https://github.com/google/cadvisor/archive/v${CADVISOR_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f cadvisor.tar.gz \
    && rm -f cadvisor.tar.gz \
    && cd cmd \
    && CGO_ENABLED=0 go build -tags netgo -ldflags="-w -s" -o cadvisor .

# Stage2: setup runtime container
FROM scratch

COPY --from=build /go/src/github.com/google/cadvisor/cmd/cadvisor /cadvisor
COPY --from=build /go/src/github.com/google/cadvisor/LICENSE /LICENSE

EXPOSE 8080

ENTRYPOINT ["/cadvisor", "-logtostderr"]
