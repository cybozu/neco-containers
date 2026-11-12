# serf container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.27-resolute AS build
ENV GOPROXY=https://golang.flatt.tech

ARG SERF_COMMIT=fe2acb9e90d89e27409a88d086553a2ad429cb8b

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p /go/src/github.com/hashicorp/serf \
&& curl -fsSL https://github.com/hashicorp/serf/archive/${SERF_COMMIT}.tar.gz | \
tar -x -z -f - --strip-components 1 -C /go/src/github.com/hashicorp/serf

WORKDIR /go/src/github.com/hashicorp/serf

RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    go install -ldflags="-w -s" ./...

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:26.04.20261111@sha256:4c899d982515115b4de5821fbd6a2208fec224244b2d6e9bb1afc396d199632f
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
