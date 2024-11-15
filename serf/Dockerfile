# serf container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.23-noble AS build

ARG SERF_VERSION=0.10.1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p /go/src/github.com/hashicorp/serf \
&& curl -fsSL https://github.com/hashicorp/serf/archive/v${SERF_VERSION}.tar.gz | \
tar -x -z -f - --strip-components 1 -C /go/src/github.com/hashicorp/serf

WORKDIR /go/src/github.com/hashicorp/serf

RUN go install -ldflags="-w -s" ./...

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

RUN apt-get update \
  && apt-get -y install --no-install-recommends \
  python3 \
  python3-requests \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /go/bin /usr/local/serf/bin
COPY --from=build /go/src/github.com/hashicorp/serf/LICENSE /usr/local/serf/LICENSE
COPY install-tools /usr/local/serf/install-tools

ENV PATH=/usr/local/serf/bin:"$PATH"

USER 10000:10000
EXPOSE 7373 7946

ENTRYPOINT ["/usr/local/serf/bin/serf"]
