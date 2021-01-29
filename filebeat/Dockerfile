# Filebeats container image

# Stage1: build from source
FROM quay.io/cybozu/golang:1.15-focal AS builder

ARG BEATS_VERSION=7.10.2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSLf https://github.com/elastic/beats/archive/v${BEATS_VERSION}.tar.gz | \
        tar zxf - -C /work/ \
    && mv /work/beats-${BEATS_VERSION} /work/beats \
    && rm -rf /work/beats/x-pack
WORKDIR /work/beats/filebeat
RUN make CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Stage2: setup runtime container
FROM scratch

COPY --from=builder /work/beats/filebeat/filebeat /filebeat
COPY --from=builder /work/beats/LICENSE.txt /LICENSE.txt
COPY --from=builder /work/beats/licenses /licenses/

USER 10000:10000

ENTRYPOINT ["/filebeat"]
