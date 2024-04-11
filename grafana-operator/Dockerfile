# Grafana Operator container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build
ARG VERSION=5.8.1

WORKDIR /work/grafana-operator
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# The grafana-operator uses the git submodule.
# The source archive on the GitHub Releases does not contain the commit ID of submodules.
# To get submodules of the specific revision, we clone the grafana-operator instead of downloading the tar archive.
# https://github.com/grafana-operator/grafana-operator/blob/master/.gitmodules
RUN git clone --depth 1 --single-branch -b v${VERSION} https://github.com/grafana/grafana-operator.git .
RUN git submodule update --init --recursive --recommend-shallow --depth 1
RUN CGO_ENABLED=0 GO111MODULE=on go build -o /usr/local/bin/grafana-operator main.go

# Stage2: setup runtime container
# refer to : https://github.com/grafana-operator/grafana-operator/blob/master/build/Dockerfile
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

USER 10000:10000
COPY --from=build /usr/local/bin/grafana-operator /usr/local/bin/grafana-operator
COPY --from=build /work/grafana-operator/LICENSE /usr/local/grafana-operator/LICENSE

ENTRYPOINT ["/usr/local/bin/grafana-operator"]
