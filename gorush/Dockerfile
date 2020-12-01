# Gorush container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.15-bionic AS build

ARG GORUSH_VERSION="1.12.0"

WORKDIR /work
RUN git clone https://github.com/appleboy/gorush.git -b "v${GORUSH_VERSION}" . \
    && make build

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /work/release/gorush /usr/local/gorush/bin/gorush
COPY --from=build /work/LICENSE /usr/local/gorush/LICENSE

ENV PATH=/usr/local/gorush/bin:"$PATH"

USER 10000:10000

# Gorush
EXPOSE 8088

ENTRYPOINT ["/usr/local/gorush/bin/gorush"]
CMD ["-c", "/config/config.yaml"]
