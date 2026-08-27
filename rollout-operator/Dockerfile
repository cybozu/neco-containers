# tempo container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG ROLLOUT_OPERATOR_SRCREPO=grafana/rollout-operator
ARG ROLLOUT_OPERATOR_VERSION=0.39.0

WORKDIR /work
RUN git clone --depth=1 -b v${ROLLOUT_OPERATOR_VERSION} https://github.com/${ROLLOUT_OPERATOR_SRCREPO}.git rollout-operator
RUN cd rollout-operator && make rollout-operator

# Stage2: setup runtime containers
FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/rollout-operator/rollout-operator /usr/bin/rollout-operator

USER 10000:10000

ENTRYPOINT ["/usr/bin/rollout-operator"]
