# Gorush container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.15-bionic AS build

ARG GORUSH_VERSION="1.13.0.cybozu.5"

WORKDIR /work
RUN git clone https://github.com/cybozu/gorush.git -b "v${GORUSH_VERSION}" . \
    && make build

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/release/gorush /usr/local/gorush/bin/gorush
COPY --from=build /work/LICENSE /usr/local/gorush/LICENSE

ENV PATH=/usr/local/gorush/bin:"$PATH"

USER 10000:10000

# Gorush
EXPOSE 8088

ENTRYPOINT ["/usr/local/gorush/bin/gorush"]
CMD ["-c", "/config/config.yaml"]
