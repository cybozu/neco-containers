FROM quay.io/cybozu/golang:1.18-jammy as build
ARG POMERIUM_VERSION=0.20.0

WORKDIR /work/pomerium
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/pomerium/pomerium/archive/v${POMERIUM_VERSION}.tar.gz | \
    tar zxf - --strip-components 1 -C /work/pomerium

RUN curl -sSLf https://deb.nodesource.com/setup_16.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global yarn
RUN make yarn
RUN make build-ui

RUN apt-get update \
    && apt-get -y --no-install-recommends install zip
RUN go mod download
RUN make build-go NAME=pomerium
RUN touch /config.yaml

FROM quay.io/cybozu/ubuntu:22.04
WORKDIR /pomerium
COPY --from=build /work/pomerium/bin/* /bin/
COPY --from=build /config.yaml ./config.yaml
COPY --from=build /work/pomerium/LICENSE ./LICENSE
ENTRYPOINT [ "/bin/pomerium" ]
CMD ["-config","/pomerium/config.yaml"]
