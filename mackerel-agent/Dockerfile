FROM quay.io/cybozu/golang:1.15-bionic as build
ARG MACKEREL_AGENT_VERSION=0.69.3

RUN curl -sLfO https://github.com/mackerelio/mackerel-agent/archive/v${MACKEREL_AGENT_VERSION}.tar.gz \
    && tar xzf v${MACKEREL_AGENT_VERSION}.tar.gz \
    && mv mackerel-agent-${MACKEREL_AGENT_VERSION} mackerel-agent
WORKDIR /work/mackerel-agent
RUN go build -o build/mackerel-agent -ldflags="-w -s -X github.com/mackerelio/mackerel-agent/version.VERSION=${MACKEREL_AGENT_VERSION}" .

FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /work/mackerel-agent/build/mackerel-agent /usr/local/bin/mackerel-agent
COPY --from=build /work/mackerel-agent/LICENSE /usr/local/bin/LICENSE
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        iproute2 \
        net-tools \
    && rm -rf /var/lib/apt/lists/*
USER 10000:10000

ENTRYPOINT ["/usr/local/bin/mackerel-agent", "--conf=/etc/mackerel-agent/mackerel-agent.conf"]
