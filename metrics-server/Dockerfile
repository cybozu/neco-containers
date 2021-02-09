# Build stage
FROM quay.io/cybozu/golang:1.15-focal AS build

ARG METRICS_SERVER_VERSION=0.4.2
# Need to clone with git because tar-ball does not contain `.git` directory and
# `.git` is needed in the building sequence.
RUN git clone https://github.com/kubernetes-sigs/metrics-server

WORKDIR /work/metrics-server
RUN git checkout v${METRICS_SERVER_VERSION} \
    && make all

# Runtime stage
FROM scratch
COPY --from=build /work/metrics-server/metrics-server /metrics-server
COPY --from=build /work/metrics-server/LICENSE /LICENSE

USER 65534:65534
EXPOSE 4443
ENTRYPOINT ["/metrics-server"]
