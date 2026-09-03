FROM ghcr.io/cybozu/golang:1.26-noble AS build

ARG FALCOSIDEKICK_VERSION=2.34.1
ARG FALCOSIDEKICK_COMMIT=03364bc26b2262439a5ddeb57e1e354742727e11 # 2.34.1

WORKDIR /work/falcosidekick
RUN curl -fsSL -o /tmp/falcosidekick-${FALCOSIDEKICK_COMMIT}.tar.gz https://github.com/falcosecurity/falcosidekick/archive/${FALCOSIDEKICK_COMMIT}.tar.gz && \
    tar -xzf /tmp/falcosidekick-${FALCOSIDEKICK_COMMIT}.tar.gz --strip-components=1 && \
    rm /tmp/falcosidekick-${FALCOSIDEKICK_COMMIT}.tar.gz && \
    CGO_ENABLED=0 go build -trimpath -ldflags \
        "-s -w -X main.GitVersion=${FALCOSIDEKICK_VERSION} -X main.gitCommit=${FALCOSIDEKICK_COMMIT} -X main.gitTreeState=clean" \
        -o /usr/local/bin/falcosidekick .

FROM ghcr.io/cybozu/ubuntu:24.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /usr/local/bin/falcosidekick /usr/bin/falcosidekick
COPY --from=build /work/falcosidekick/LICENSE /usr/local/falcosidekick/LICENSE

USER 10000:10000
EXPOSE 2801

ENTRYPOINT ["/usr/bin/falcosidekick"]
