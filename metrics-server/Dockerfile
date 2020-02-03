# Build stage
FROM quay.io/cybozu/golang:1.13-bionic AS build

ARG METRICS_SERVER_VERSION=0.3.6
# Use workdir `kubernetes-incubator` because `kubernetes-incubator` is used for
# import paths instead of `kubernetes-sigs` in metrics-server.
WORKDIR /go/src/github.com/kubernetes-incubator
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# Need to clone with git because tar-ball does not contain `.git` directory and
# `.git` is needed in the building sequence.
RUN git clone https://github.com/kubernetes-sigs/metrics-server && \
    cd metrics-server && \
    git checkout v${METRICS_SERVER_VERSION} && \
    make all

# Runtime stage
FROM scratch
COPY --from=build /go/src/github.com/kubernetes-incubator/metrics-server/_output/amd64/metrics-server /usr/local/metrics-server/bin/metrics-server
COPY --from=build /go/src/github.com/kubernetes-incubator/metrics-server/LICENSE /usr/local/metrics-server/bin/LICENSE
ENV PATH=/usr/local/metrics-server/bin:"$PATH"

USER 65534:65534
EXPOSE 4443
ENTRYPOINT ["metrics-server"]
