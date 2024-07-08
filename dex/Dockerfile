# dex container image

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG DEX_VERSION=2.38.0

WORKDIR /work/dex/
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/dexidp/dex/archive/v${DEX_VERSION}.tar.gz | \
    tar zxf - -C /work/dex --strip-components 1

RUN make bin/dex VERSION=v${DEX_VERSION}

# Stage2: setup runtime container
FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/dex/bin/dex /usr/local/bin/dex
COPY --from=build /work/dex/LICENSE /usr/local/share/doc/dex/LICENSE
COPY --from=build /work/dex/web /web

USER 10000:10000

# the lines below follow the old upstream Dockerfile: https://github.com/dexidp/dex/blob/2dccdc2a1a4f5cb6562b54225a8615280e5ac109/Dockerfile
# if you want to use this outside of Neco, please edit ENTRYPOINT and COMMAND apporpriately
ENTRYPOINT ["dex"]
CMD ["version"]
