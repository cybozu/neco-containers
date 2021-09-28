# consul container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.17-focal AS build

ARG CONSUL_VERSION=1.10.2
ARG DOCKER_CONSUL_REVISION=d7b8436b3c10524e018be6676a80fe872fbfa891

WORKDIR /work
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN git clone --depth=1 -b v${CONSUL_VERSION} https://github.com/hashicorp/consul /work/consul

# docker-entrypoint.sh contains non- bourne shell compatible variable substitution
RUN curl -sSfL -o docker-entrypoint.sh https://raw.githubusercontent.com/hashicorp/docker-consul/${DOCKER_CONSUL_REVISION}/0.X/docker-entrypoint.sh \
    && sed -i -e '1s,/bin/sh,/bin/bash,' docker-entrypoint.sh \
    && chmod +x docker-entrypoint.sh
RUN curl -sSfL -o LICENSE-docker-consul https://raw.githubusercontent.com/hashicorp/docker-consul/${DOCKER_CONSUL_REVISION}/LICENSE

WORKDIR /work/consul
RUN CGO_ENABLED=0 go install -ldflags="-w -s"

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:20.04

RUN apt-get update \
    && apt-get -y install --no-install-recommends dumb-init \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /go/bin/consul /bin/consul
COPY --from=build /work/consul/LICENSE /LICENSE
COPY --from=build /work/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
COPY --from=build /work/LICENSE-docker-consul /LICENSE-docker-consul
RUN mkdir -p /consul/data && \
    mkdir -p /consul/config && \
    chown -R 10000:10000 /consul

USER 10000:10000
VOLUME /consul/data
EXPOSE 8300
EXPOSE 8301 8301/udp 8302 8302/udp
EXPOSE 8500 8600 8600/udp
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["agent", "-dev", "-client", "0.0.0.0"]
