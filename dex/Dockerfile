# dex container image

# Stage1: build from source
FROM quay.io/cybozu/golang:1.15-bionic AS build

ARG DEX_VERSION=2.27.0

WORKDIR /work/dex/
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/dexidp/dex/archive/v${DEX_VERSION}.tar.gz | \
        tar zxf - -C /work/dex --strip-components 1

RUN make bin/dex VERSION=v${DEX_VERSION}

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /work/dex/bin/dex /usr/local/bin/dex
COPY --from=build /work/dex/LICENSE /usr/local/share/doc/dex/LICENSE
COPY --from=build /work/dex/web /web

USER 10000:10000

ENTRYPOINT ["dex"]
CMD ["version"]
